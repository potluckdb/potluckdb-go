// Package potluckdb is the Go facade over the potluckdb Native Application
// ABI: the canonical 22-intent API with validated policies, structured
// errors, and finalizer-guarded owner lifecycle. Link the packaged
// potluckdb_runtime_ffi library from the Universal C SDK.
package potluckdb

/*
#cgo CFLAGS: -I${SRCDIR}/include -I${SRCDIR}/../../c-sdk/include -I${SRCDIR}/../../interface/include
#cgo LDFLAGS: -lpotluckdb_runtime_ffi

#include <stdlib.h>
#include <string.h>
#include "potluckdb.h"
*/
import "C"

import (
	"encoding/hex"
	"fmt"
	"regexp"
	"runtime"
	"sort"
	"sync"
	"time"
	"unsafe"
)

// ErrorCategory classifies a failure the same way every potluckdb SDK does.
type ErrorCategory int

const (
	CategoryCaller    ErrorCategory = 1
	CategoryTemporary ErrorCategory = 2
	CategoryPermanent ErrorCategory = 3
	CategoryInternal  ErrorCategory = 4
)

// Error is the structured potluckdb application error.
type Error struct {
	Code       string
	Category   ErrorCategory
	Message    string
	Action     string
	RetryAfter *time.Duration
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func invalidArgument(message, action string) *Error {
	return &Error{
		Code:     "invalid_argument",
		Category: CategoryCaller,
		Message:  message,
		Action:   action,
	}
}

func internalFailure(message string) *Error {
	return &Error{
		Code:     "internal_failure",
		Category: CategoryInternal,
		Message:  message,
		Action:   "Preserve the failure details and report this error.",
	}
}

func bufferString(buffer C.PotluckDbBufferV1) string {
	if buffer.data == nil || buffer.length == 0 {
		return ""
	}
	return C.GoStringN((*C.char)(unsafe.Pointer(buffer.data)), C.int(buffer.length))
}

func sliceBytes(slice C.PotluckDbSliceV1) []byte {
	if slice.data == nil || slice.length == 0 {
		return []byte{}
	}
	return C.GoBytes(unsafe.Pointer(slice.data), C.int(slice.length))
}

func check(result C.int32_t, cError *C.PotluckDbErrorV1) error {
	if result == 0 {
		C.potluckdb_error_release_v1(cError)
		return nil
	}
	category := ErrorCategory(cError.category)
	if category < CategoryCaller || category > CategoryInternal {
		category = CategoryInternal
	}
	failure := &Error{
		Code:     bufferString(cError.code),
		Category: category,
		Message:  bufferString(cError.message),
		Action:   bufferString(cError.action),
	}
	if cError.has_retry_after != 0 {
		retry := time.Duration(cError.retry_after_ms) * time.Millisecond
		failure.RetryAfter = &retry
	}
	C.potluckdb_error_release_v1(cError)
	return failure
}

func withUtf8(value string, body func(C.PotluckDbSliceV1) error) error {
	if len(value) == 0 {
		return body(C.PotluckDbSliceV1{})
	}
	bytes := []byte(value)
	slice := C.PotluckDbSliceV1{
		data:   (*C.uint8_t)(unsafe.Pointer(&bytes[0])),
		length: C.uint64_t(len(bytes)),
	}
	err := body(slice)
	runtime.KeepAlive(bytes)
	return err
}

// Revision identifies one published database state.
type Revision struct {
	bytes [32]byte
}

// NewRevision copies an exact 32-byte revision identity.
func NewRevision(value []byte) (*Revision, error) {
	if len(value) != 32 {
		return nil, invalidArgument(
			"revision must contain exactly 32 bytes",
			"Supply a 32-byte revision and retry.",
		)
	}
	revision := &Revision{}
	copy(revision.bytes[:], value)
	return revision, nil
}

// ParseRevision parses the canonical printed revision form.
func ParseRevision(text string) (*Revision, error) {
	var raw C.PotluckDbRevisionV1
	err := withUtf8(text, func(slice C.PotluckDbSliceV1) error {
		cError := newError()
		return check(C.potluckdb_revision_parse_v1(slice, &raw, &cError), &cError)
	})
	if err != nil {
		return nil, err
	}
	return readRevision(raw), nil
}

// Formatted renders the canonical printed revision form.
func (r *Revision) Formatted() (string, error) {
	raw := writeRevision(r)
	var text C.PotluckDbBufferV1
	cError := newError()
	if err := check(
		C.potluckdb_revision_format_v1(&raw, &text, &cError), &cError,
	); err != nil {
		return "", err
	}
	value := bufferString(text)
	C.potluckdb_buffer_release_v1(&text)
	return value, nil
}

// Bytes copies the raw 32-byte identity.
func (r *Revision) Bytes() []byte {
	value := make([]byte, 32)
	copy(value, r.bytes[:])
	return value
}

// Equal reports whether both revisions carry the same identity.
func (r *Revision) Equal(other *Revision) bool {
	return other != nil && r.bytes == other.bytes
}

func (r *Revision) String() string {
	return hex.EncodeToString(r.bytes[:])
}

func readRevision(raw C.PotluckDbRevisionV1) *Revision {
	revision := &Revision{}
	C.memcpy(
		unsafe.Pointer(&revision.bytes[0]), unsafe.Pointer(&raw.bytes[0]), 32,
	)
	return revision
}

func writeRevision(revision *Revision) C.PotluckDbRevisionV1 {
	var raw C.PotluckDbRevisionV1
	C.memcpy(
		unsafe.Pointer(&raw.bytes[0]), unsafe.Pointer(&revision.bytes[0]), 32,
	)
	return raw
}

func newError() C.PotluckDbErrorV1 {
	var cError C.PotluckDbErrorV1
	C.potluckdb_error_init_v1(&cError)
	return cError
}

// TransactionMode selects the SQLite transaction acquisition strategy.
type TransactionMode int

const (
	Deferred  TransactionMode = 1
	Immediate TransactionMode = 2
	Exclusive TransactionMode = 3
)

// AutomaticSyncOptions tunes the shared automatic synchronization driver.
type AutomaticSyncOptions struct {
	OutboundDebounce   time.Duration
	ActivePullInterval time.Duration
	IdlePullInterval   time.Duration
	RetryInitial       time.Duration
	RetryMaximum       time.Duration
}

// DefaultAutomaticSyncOptions returns the canonical automatic-sync timers.
func DefaultAutomaticSyncOptions() AutomaticSyncOptions {
	return AutomaticSyncOptions{
		OutboundDebounce:   30 * time.Millisecond,
		ActivePullInterval: 200 * time.Millisecond,
		IdlePullInterval:   time.Second,
		RetryInitial:       250 * time.Millisecond,
		RetryMaximum:       30 * time.Second,
	}
}

// SyncPolicy selects automatic or manual synchronization.
type SyncPolicy struct {
	automatic bool
	options   AutomaticSyncOptions
}

// AutomaticSync returns the default automatic policy.
func AutomaticSync() *SyncPolicy {
	return &SyncPolicy{automatic: true, options: DefaultAutomaticSyncOptions()}
}

// AutomaticSyncWith validates and applies explicit automatic-sync timers.
func AutomaticSyncWith(options AutomaticSyncOptions) (*SyncPolicy, error) {
	maximum := 24 * time.Hour
	timers := []time.Duration{
		options.OutboundDebounce,
		options.ActivePullInterval,
		options.IdlePullInterval,
		options.RetryInitial,
		options.RetryMaximum,
	}
	for _, timer := range timers {
		if timer < time.Millisecond || timer > maximum {
			return nil, invalidArgument(
				"sync timers must be from 1 through 86400000 milliseconds",
				"Correct the timer values and retry.",
			)
		}
	}
	if options.IdlePullInterval < options.ActivePullInterval {
		return nil, invalidArgument(
			"idle pull interval must be at least the active pull interval",
			"Correct the pull intervals and retry.",
		)
	}
	if options.RetryMaximum < options.RetryInitial {
		return nil, invalidArgument(
			"maximum retry delay must be at least the initial retry delay",
			"Correct the retry delays and retry.",
		)
	}
	return &SyncPolicy{automatic: true, options: options}, nil
}

// ManualSync returns the manual policy: only Sync() moves data.
func ManualSync() *SyncPolicy {
	return &SyncPolicy{automatic: false, options: DefaultAutomaticSyncOptions()}
}

// PrefetchPolicy tunes read-ahead for connections.
type PrefetchPolicy struct {
	enabled               bool
	backgroundTargetBytes uint64
	memoryBudgetBytes     uint64
}

// AutomaticPrefetch returns the default prefetch policy.
func AutomaticPrefetch() *PrefetchPolicy {
	return &PrefetchPolicy{enabled: true, memoryBudgetBytes: 64 << 20}
}

// AutomaticPrefetchWith validates explicit prefetch bounds.
func AutomaticPrefetchWith(
	backgroundTargetBytes, memoryBudgetBytes uint64,
) (*PrefetchPolicy, error) {
	if memoryBudgetBytes == 0 {
		return nil, invalidArgument(
			"prefetch memory budget must be positive",
			"Raise the prefetch memory budget and retry.",
		)
	}
	return &PrefetchPolicy{
		enabled:               true,
		backgroundTargetBytes: backgroundTargetBytes,
		memoryBudgetBytes:     memoryBudgetBytes,
	}, nil
}

// DisabledPrefetch turns read-ahead off.
func DisabledPrefetch() *PrefetchPolicy {
	return &PrefetchPolicy{}
}

// CheckpointPolicy tunes WAL checkpointing for connections.
type CheckpointPolicy struct {
	enabled          bool
	softLimitBytes   uint64
	forcedLimitBytes uint64
}

// AutomaticCheckpoint returns the default checkpoint policy.
func AutomaticCheckpoint() *CheckpointPolicy {
	return &CheckpointPolicy{
		enabled:          true,
		softLimitBytes:   64 << 20,
		forcedLimitBytes: 128 << 20,
	}
}

// AutomaticCheckpointWith validates explicit checkpoint limits.
func AutomaticCheckpointWith(
	softLimitBytes, forcedLimitBytes uint64,
) (*CheckpointPolicy, error) {
	if softLimitBytes == 0 || forcedLimitBytes < softLimitBytes {
		return nil, invalidArgument(
			"checkpoint limits must be positive and forced must not be below soft",
			"Correct the checkpoint limits and retry.",
		)
	}
	return &CheckpointPolicy{
		enabled:          true,
		softLimitBytes:   softLimitBytes,
		forcedLimitBytes: forcedLimitBytes,
	}, nil
}

// DisabledCheckpoint turns automatic checkpointing off.
func DisabledCheckpoint() *CheckpointPolicy {
	return &CheckpointPolicy{}
}

// RetentionPolicy bounds retained revisions during collection.
type RetentionPolicy struct {
	keepLast  uint64
	collectAt uint64
}

// BoundedRetention validates a keep/collect pair; zeros select the defaults.
func BoundedRetention(keepLast, collectAt uint64) (*RetentionPolicy, error) {
	if keepLast == 0 && collectAt == 0 {
		keepLast, collectAt = 10, 20
	}
	if keepLast == 0 || collectAt <= keepLast {
		return nil, invalidArgument(
			"retention keepLast must be positive and collectAt must be greater",
			"Correct the retention bounds and retry.",
		)
	}
	return &RetentionPolicy{keepLast: keepLast, collectAt: collectAt}, nil
}

// ConnectionOptions tunes one connection.
type ConnectionOptions struct {
	Timeout     time.Duration
	MaximumRows uint64
	Prefetch    *PrefetchPolicy
	Checkpoint  *CheckpointPolicy
}

// DefaultConnectionOptions returns the canonical connection settings.
func DefaultConnectionOptions() *ConnectionOptions {
	return &ConnectionOptions{
		Timeout:     5 * time.Second,
		MaximumRows: 100_000,
		Prefetch:    AutomaticPrefetch(),
		Checkpoint:  AutomaticCheckpoint(),
	}
}

func (o *ConnectionOptions) normalized() (*ConnectionOptions, error) {
	if o == nil {
		return DefaultConnectionOptions(), nil
	}
	if o.Timeout > 24*time.Hour || o.MaximumRows == 0 {
		return nil, invalidArgument(
			"connection timeout or result limit is outside the supported bounds",
			"Use a timeout up to one day and a positive MaximumRows.",
		)
	}
	value := *o
	if value.Prefetch == nil {
		value.Prefetch = AutomaticPrefetch()
	}
	if value.Checkpoint == nil {
		value.Checkpoint = AutomaticCheckpoint()
	}
	return &value, nil
}

var clientIDShape = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._:-]{0,127}$`)

// OpenOptions tunes one Database or Snapshot session.
type OpenOptions struct {
	LocalBudgetBytes        uint64
	MinimumLocalBudgetBytes uint64
	LocalCompactionBytes    uint64
	EncryptionKey           []byte
	SyncPolicy              *SyncPolicy
	ClientID                string
	Connection              *ConnectionOptions
}

// DefaultOpenOptions returns the canonical open settings.
func DefaultOpenOptions() *OpenOptions {
	return &OpenOptions{
		LocalBudgetBytes:        1 << 30,
		MinimumLocalBudgetBytes: 10 << 20,
		LocalCompactionBytes:    256 << 20,
		SyncPolicy:              AutomaticSync(),
		Connection:              DefaultConnectionOptions(),
	}
}

func (o *OpenOptions) normalized() (*OpenOptions, error) {
	if o == nil {
		return DefaultOpenOptions(), nil
	}
	if o.LocalBudgetBytes < o.MinimumLocalBudgetBytes {
		return nil, invalidArgument(
			"local budget must be at least the minimum local budget",
			"Raise the local budget or lower the minimum and retry.",
		)
	}
	if o.LocalCompactionBytes == 0 {
		return nil, invalidArgument(
			"local compaction threshold must be positive",
			"Raise the compaction threshold and retry.",
		)
	}
	if o.EncryptionKey != nil && len(o.EncryptionKey) != 32 {
		return nil, invalidArgument(
			"encryption key must contain exactly 32 bytes",
			"Supply a 32-byte key and retry.",
		)
	}
	if o.ClientID != "" && !clientIDShape.MatchString(o.ClientID) {
		return nil, invalidArgument(
			"ClientID must be a portable identity of 1 through 128 characters",
			"Correct the client identity and retry.",
		)
	}
	value := *o
	if value.SyncPolicy == nil {
		value.SyncPolicy = AutomaticSync()
	}
	connection, err := value.Connection.normalized()
	if err != nil {
		return nil, err
	}
	value.Connection = connection
	return &value, nil
}

// RemoteOptions tunes one Remote.
type RemoteOptions struct {
	ReadConcurrency  uint32
	WriteConcurrency uint32
	Retention        *RetentionPolicy
}

// DefaultRemoteOptions returns the canonical remote settings.
func DefaultRemoteOptions() *RemoteOptions {
	return &RemoteOptions{ReadConcurrency: 32, WriteConcurrency: 16}
}

func (o *RemoteOptions) normalized() (*RemoteOptions, error) {
	if o == nil {
		return DefaultRemoteOptions(), nil
	}
	if o.ReadConcurrency < 1 || o.ReadConcurrency > 256 ||
		o.WriteConcurrency < 1 || o.WriteConcurrency > 128 {
		return nil, invalidArgument(
			"remote concurrency is outside the supported bounds",
			"Use read concurrency 1 through 256 and write concurrency 1 through 128.",
		)
	}
	value := *o
	return &value, nil
}

// Bindings carries SQL parameters. Values must be nil, bool, int, int64,
// float64, string, or []byte.
type Bindings struct {
	kind    C.uint32_t
	entries []bindingEntry
}

type bindingEntry struct {
	name  string
	value any
}

// Positional binds values by position.
func Positional(values ...any) Bindings {
	entries := make([]bindingEntry, len(values))
	for index, value := range values {
		entries[index] = bindingEntry{value: value}
	}
	return Bindings{kind: 1, entries: entries}
}

// Named binds values by parameter name.
func Named(values map[string]any) Bindings {
	names := make([]string, 0, len(values))
	for name := range values {
		names = append(names, name)
	}
	sort.Strings(names)
	entries := make([]bindingEntry, len(names))
	for index, name := range names {
		entries[index] = bindingEntry{name: name, value: values[name]}
	}
	return Bindings{kind: 2, entries: entries}
}

func noBindings() Bindings {
	return Bindings{kind: 1}
}

func buildBindings(
	bindings Bindings, body func(*C.PotluckDbBindingsV1) error,
) error {
	var raw C.PotluckDbBindingsV1
	C.potluckdb_bindings_init_v1(&raw)
	raw.kind = bindings.kind
	count := len(bindings.entries)
	if count == 0 {
		return body(&raw)
	}
	// The item array and every bound buffer live in C memory for the call, so
	// no Go pointer is reachable from the native binding structures.
	size := C.size_t(count) * C.size_t(unsafe.Sizeof(C.PotluckDbBindingV1{}))
	items := (*C.PotluckDbBindingV1)(C.malloc(size))
	if items == nil {
		return internalFailure("binding allocation failed")
	}
	var allocations []unsafe.Pointer
	defer func() {
		for _, pointer := range allocations {
			C.free(pointer)
		}
		C.free(unsafe.Pointer(items))
	}()
	cSlice := func(bytes []byte) C.PotluckDbSliceV1 {
		if len(bytes) == 0 {
			return C.PotluckDbSliceV1{}
		}
		pointer := C.CBytes(bytes)
		allocations = append(allocations, pointer)
		return C.PotluckDbSliceV1{
			data:   (*C.uint8_t)(pointer),
			length: C.uint64_t(len(bytes)),
		}
	}
	view := unsafe.Slice(items, count)
	for index, entry := range bindings.entries {
		C.potluckdb_binding_init_v1(&view[index])
		if entry.name != "" {
			view[index].name = cSlice([]byte(entry.name))
		}
		C.potluckdb_value_init_v1(&view[index].value)
		switch value := entry.value.(type) {
		case nil:
			view[index].value.kind = 1
		case bool:
			view[index].value.kind = 2
			if value {
				view[index].value.boolean_value = 1
			}
		case int:
			view[index].value.kind = 3
			view[index].value.integer_value = C.int64_t(value)
		case int64:
			view[index].value.kind = 3
			view[index].value.integer_value = C.int64_t(value)
		case float64:
			view[index].value.kind = 4
			view[index].value.real_value = C.double(value)
		case string:
			view[index].value.kind = 5
			view[index].value.bytes = cSlice([]byte(value))
		case []byte:
			view[index].value.kind = 6
			view[index].value.bytes = cSlice(value)
		default:
			return invalidArgument(
				"binding values must be nil, bool, int, int64, float64, string, or []byte",
				"Convert the value to a supported SQL type and retry.",
			)
		}
	}
	raw.items = items
	raw.count = C.uint64_t(count)
	return body(&raw)
}

func copyRows(handle C.PotluckDbRowsHandleV1) ([][]any, error) {
	defer C.potluckdb_rows_release_v1(handle, nil)
	cError := newError()
	var rowCount C.uint64_t
	if err := check(
		C.potluckdb_rows_row_count_v1(handle, &rowCount, &cError), &cError,
	); err != nil {
		return nil, err
	}
	rows := make([][]any, 0, int(rowCount))
	for row := C.uint64_t(0); row < rowCount; row++ {
		cError = newError()
		var columnCount C.uint64_t
		if err := check(
			C.potluckdb_rows_column_count_v1(handle, row, &columnCount, &cError),
			&cError,
		); err != nil {
			return nil, err
		}
		values := make([]any, 0, int(columnCount))
		for column := C.uint64_t(0); column < columnCount; column++ {
			var value C.PotluckDbValueV1
			C.potluckdb_value_init_v1(&value)
			cError = newError()
			if err := check(
				C.potluckdb_rows_value_v1(
					handle,
					row,
					column,
					&value,
					C.uint32_t(unsafe.Sizeof(value)),
					&cError,
				),
				&cError,
			); err != nil {
				return nil, err
			}
			switch value.kind {
			case 1:
				values = append(values, nil)
			case 2:
				values = append(values, value.boolean_value != 0)
			case 3:
				values = append(values, int64(value.integer_value))
			case 4:
				values = append(values, float64(value.real_value))
			case 5:
				values = append(values, string(sliceBytes(value.bytes)))
			case 6:
				values = append(values, sliceBytes(value.bytes))
			default:
				return nil, internalFailure(
					"potluckdb returned an unknown SQL value kind",
				)
			}
		}
		rows = append(rows, values)
	}
	return rows, nil
}

func copyStrings(handle C.PotluckDbStringListHandleV1) ([]string, error) {
	cError := newError()
	var count C.uint64_t
	if err := check(
		C.potluckdb_string_list_count_v1(handle, &count, &cError), &cError,
	); err != nil {
		return nil, err
	}
	values := make([]string, 0, int(count))
	for index := C.uint64_t(0); index < count; index++ {
		var slice C.PotluckDbSliceV1
		cError = newError()
		if err := check(
			C.potluckdb_string_list_get_v1(handle, index, &slice, &cError), &cError,
		); err != nil {
			return nil, err
		}
		values = append(values, string(sliceBytes(slice)))
	}
	return values, nil
}

// ExecuteResult reports one write statement's effect.
type ExecuteResult struct {
	RowsAffected int64
	LastInsertID *int64
}

// SyncResult is one of UpToDate, Pushed, Pulled, or Conflict.
type SyncResult interface {
	syncResult()
}

// UpToDate reports no transferable changes.
type UpToDate struct {
	Revision *Revision
}

// Pushed reports locally published changes.
type Pushed struct {
	Revision *Revision
}

// Pulled reports replayed remote changes.
type Pulled struct {
	Revision    *Revision
	Replayed    []string
	Quarantined []string
}

// Conflict reports an unresolved divergence.
type Conflict struct {
	RemoteRevision *Revision
	Reason         string
}

func (UpToDate) syncResult() {}
func (Pushed) syncResult()   {}
func (Pulled) syncResult()   {}
func (Conflict) syncResult() {}

// ConflictResolution reports the outcome of resolving a conflict.
type ConflictResolution struct {
	Revision         *Revision
	AlreadyPublished []string
	Replayed         []string
	Quarantined      []string
	Pending          []string
}

// CollectionReport summarizes one retention collection pass.
type CollectionReport struct {
	FloorBefore *Revision
	FloorAfter  *Revision
	Epoch       uint64
	Marked      uint64
	Swept       uint64
	Aborted     bool
}

// DatabaseStatus is the Database observational status.
type DatabaseStatus struct {
	Revision               *Revision
	PendingWrites          uint64
	LocalBytes             uint64
	LocalBudgetBytes       uint64
	ConnectionCount        uint64
	SyncState              string
	NextSyncIn             *time.Duration
	ConflictRemoteRevision *Revision
	Closed                 bool
}

// CheckpointStatus is one connection's checkpoint view.
type CheckpointStatus struct {
	SoftLimitBytes           uint64
	ForcedLimitBytes         uint64
	WalBytes                 uint64
	Due                      bool
	Scheduled                bool
	SoftScheduled            uint64
	ForcedCheckpointAttempts uint64
	Manual                   uint64
	Completed                uint64
	Failures                 uint64
	StaleCommands            uint64
	PagesVerified            uint64
	WalBytesReclaimed        uint64
}

// PrefetchStatus is one connection's prefetch view.
type PrefetchStatus struct {
	BackgroundState       string
	Limit                 uint64
	Active                uint64
	Queued                uint64
	ReadyPages            uint64
	ReadyBytes            uint64
	BackgroundTargetBytes uint64
	PolicyFailures        uint64
}

// ConnectionStatus is the Connection observational status.
type ConnectionStatus struct {
	TransactionActive bool
	Checkpoint        *CheckpointStatus
	Prefetch          *PrefetchStatus
	Closed            bool
}

// SnapshotStatus is the Snapshot observational status.
type SnapshotStatus struct {
	Revision   *Revision
	LocalBytes uint64
	Closed     bool
}

// RemoteStatus is the Remote observational status.
type RemoteStatus struct {
	RetainedRevisions  *uint64
	MaintenancePending bool
	PolicyFailures     uint64
	Closed             bool
}

// CheckpointResult reports one manual checkpoint.
type CheckpointResult struct {
	LogFrames          int32
	CheckpointedFrames int32
}

// KeyRotationReport reports one re-encryption between remotes.
type KeyRotationReport struct {
	SourceRevision *Revision
	TargetRevision *Revision
	SourceKeyID    *Revision
	TargetKeyID    *Revision
	PieceCount     uint64
	Published      bool
}

func syncStateName(value C.uint32_t) string {
	switch value {
	case 1:
		return "manual"
	case 2:
		return "unbound"
	case 3:
		return "scheduled"
	case 4:
		return "synchronizing"
	case 5:
		return "idle"
	case 6:
		return "backoff"
	case 7:
		return "conflict"
	case 8:
		return "failed"
	case 9:
		return "closed"
	default:
		return "unknown"
	}
}

func backgroundStateName(value C.uint32_t) string {
	switch value {
	case 0:
		return "disabled"
	case 1:
		return "idle"
	case 2:
		return "fetching"
	case 3:
		return "installing"
	case 4:
		return "paused"
	case 5:
		return "complete"
	case 6:
		return "saturated"
	case 7:
		return "failed"
	case 8:
		return "closed"
	default:
		return "unknown"
	}
}

const (
	ownerRemote      = 1
	ownerDatabase    = 2
	ownerConnection  = 3
	ownerSnapshot    = 4
	ownerTransaction = 5
)

type handleState struct {
	mutex    sync.Mutex
	owner    int
	handle   C.uint64_t
	closable bool
	closed   bool
	retained any
}

func newHandleState(
	owner int, handle C.uint64_t, closable bool, retained any,
) (*handleState, error) {
	if handle == 0 {
		return nil, internalFailure("potluckdb returned a zero owner handle")
	}
	state := &handleState{
		owner:    owner,
		handle:   handle,
		closable: closable,
		retained: retained,
	}
	runtime.SetFinalizer(state, (*handleState).cleanup)
	return state, nil
}

func (s *handleState) withHandle(body func(C.uint64_t) error) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s.handle == 0 {
		return &Error{
			Code:     "invalid_state",
			Category: CategoryCaller,
			Message:  "potluckdb owner is released",
			Action:   "Open a new owner and retry.",
		}
	}
	return body(s.handle)
}

func (s *handleState) close(call func(C.uint64_t) error) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s.handle == 0 || s.closed {
		return nil
	}
	if err := call(s.handle); err != nil {
		return err
	}
	s.closed = true
	return nil
}

func (s *handleState) release(call func(C.uint64_t) error) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s.handle == 0 {
		return nil
	}
	if err := call(s.handle); err != nil {
		return err
	}
	s.handle = 0
	runtime.SetFinalizer(s, nil)
	return nil
}

func (s *handleState) cleanup() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s.handle == 0 {
		return
	}
	needsClose := s.closable && !s.closed
	switch s.owner {
	case ownerRemote:
		if needsClose {
			C.potluckdb_remote_close_v1(s.handle, nil)
		}
		C.potluckdb_remote_release_v1(s.handle, nil)
	case ownerDatabase:
		if needsClose {
			C.potluckdb_database_close_v1(s.handle, nil)
		}
		C.potluckdb_database_release_v1(s.handle, nil)
	case ownerConnection:
		if needsClose {
			C.potluckdb_connection_close_v1(s.handle, nil)
		}
		C.potluckdb_connection_release_v1(s.handle, nil)
	case ownerSnapshot:
		if needsClose {
			C.potluckdb_snapshot_close_v1(s.handle, nil)
		}
		C.potluckdb_snapshot_release_v1(s.handle, nil)
	default:
		C.potluckdb_transaction_release_v1(s.handle, nil)
	}
	s.handle = 0
}

var pairTieLock sync.Mutex

func withPair(
	left, right *handleState, body func(C.uint64_t, C.uint64_t) error,
) error {
	if left == right {
		return invalidArgument(
			"source and target must be distinct Remote values",
			"Open a distinct target Remote and retry.",
		)
	}
	leftAddress := uintptr(unsafe.Pointer(left))
	rightAddress := uintptr(unsafe.Pointer(right))
	locked := func() error {
		if left.handle == 0 || right.handle == 0 {
			return &Error{
				Code:     "invalid_state",
				Category: CategoryCaller,
				Message:  "potluckdb Remote is released",
				Action:   "Open new Remote values and retry.",
			}
		}
		return body(left.handle, right.handle)
	}
	if leftAddress == rightAddress {
		// Address ties cannot order the two mutexes; serialize globally.
		pairTieLock.Lock()
		defer pairTieLock.Unlock()
		left.mutex.Lock()
		defer left.mutex.Unlock()
		right.mutex.Lock()
		defer right.mutex.Unlock()
		return locked()
	}
	first, second := left, right
	if rightAddress < leftAddress {
		first, second = right, left
	}
	first.mutex.Lock()
	defer first.mutex.Unlock()
	second.mutex.Lock()
	defer second.mutex.Unlock()
	return locked()
}
