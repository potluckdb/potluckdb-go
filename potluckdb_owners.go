package potluckdb

/*
#include <string.h>

#include "potluckdb.h"
*/
import "C"

import (
	"runtime"
	"time"
	"unsafe"
)

// Remote is one storage location holding published database states.
type Remote struct {
	state *handleState
}

// RemoteFromDirectory opens a filesystem-backed Remote; nil options select
// the defaults.
func RemoteFromDirectory(path string, options *RemoteOptions) (*Remote, error) {
	selected, err := options.normalized()
	if err != nil {
		return nil, err
	}
	var handle C.PotluckDbRemoteHandleV1
	err = withUtf8(path, func(slice C.PotluckDbSliceV1) error {
		raw := buildRemoteOptions(selected)
		cError := newError()
		return check(
			C.potluckdb_remote_from_directory_v1(slice, &raw, &handle, &cError),
			&cError,
		)
	})
	if err != nil {
		return nil, err
	}
	state, err := newHandleState(ownerRemote, handle, true, nil)
	if err != nil {
		return nil, err
	}
	return &Remote{state: state}, nil
}

// StorageAdapter is the low-level escape hatch: it wraps a caller-owned
// native PotluckDbRuntimeBucketAdapterV1 whose callbacks and context must
// stay valid until the borrowing Remote is closed.
type StorageAdapter struct {
	pointer unsafe.Pointer
	owner   any
}

// RetainStorageAdapter keeps owner reachable for the adapter's lifetime.
func RetainStorageAdapter(pointer unsafe.Pointer, owner any) (*StorageAdapter, error) {
	if pointer == nil {
		return nil, invalidArgument(
			"storage adapter pointer must not be null",
			"Supply a live native bucket adapter and retry.",
		)
	}
	return &StorageAdapter{pointer: pointer, owner: owner}, nil
}

// BorrowedStorageAdapter trusts the caller to outlive the Remote.
func BorrowedStorageAdapter(pointer unsafe.Pointer) (*StorageAdapter, error) {
	return RetainStorageAdapter(pointer, nil)
}

// RemoteFromStorage opens a callback-backed Remote; nil options select the
// defaults.
func RemoteFromStorage(
	storage *StorageAdapter, options *RemoteOptions,
) (*Remote, error) {
	selected, err := options.normalized()
	if err != nil {
		return nil, err
	}
	raw := buildRemoteOptions(selected)
	var handle C.PotluckDbRemoteHandleV1
	cError := newError()
	if err := check(
		C.potluckdb_remote_from_storage_v1(
			(*C.PotluckDbRuntimeBucketAdapterV1)(storage.pointer),
			&raw,
			&handle,
			&cError,
		),
		&cError,
	); err != nil {
		return nil, err
	}
	state, err := newHandleState(ownerRemote, handle, true, storage)
	if err != nil {
		return nil, err
	}
	return &Remote{state: state}, nil
}

func buildRemoteOptions(options *RemoteOptions) C.PotluckDbRemoteOptionsV1 {
	var raw C.PotluckDbRemoteOptionsV1
	C.potluckdb_remote_options_init_v1(&raw)
	raw.read_concurrency = C.uint32_t(options.ReadConcurrency)
	raw.write_concurrency = C.uint32_t(options.WriteConcurrency)
	if options.Retention != nil {
		raw.flags |= 1
		C.potluckdb_retention_policy_init_v1(&raw.retention)
		raw.retention.keep_last = C.uint64_t(options.Retention.keepLast)
		raw.retention.collect_at = C.uint64_t(options.Retention.collectAt)
	}
	return raw
}

// Status reports the Remote observational status.
func (r *Remote) Status() (*RemoteStatus, error) {
	var status *RemoteStatus
	err := r.state.withHandle(func(handle C.uint64_t) error {
		var raw C.PotluckDbRemoteStatusV1
		C.potluckdb_remote_status_init_v1(&raw)
		cError := newError()
		if err := check(
			C.potluckdb_remote_status_v1(
				handle, &raw, C.uint32_t(unsafe.Sizeof(raw)), &cError,
			),
			&cError,
		); err != nil {
			return err
		}
		status = &RemoteStatus{
			MaintenancePending: raw.maintenance_pending != 0,
			PolicyFailures:     uint64(raw.policy_failures),
			Closed:             raw.closed != 0,
		}
		if raw.has_retained_revisions != 0 {
			retained := uint64(raw.retained_revisions)
			status.RetainedRevisions = &retained
		}
		return nil
	})
	return status, err
}

// Collect runs one bounded retention collection; nil selects the defaults.
func (r *Remote) Collect(retention *RetentionPolicy) (*CollectionReport, error) {
	selected := retention
	if selected == nil {
		bounded, err := BoundedRetention(0, 0)
		if err != nil {
			return nil, err
		}
		selected = bounded
	}
	var report *CollectionReport
	err := r.state.withHandle(func(handle C.uint64_t) error {
		var raw C.PotluckDbRetentionPolicyV1
		C.potluckdb_retention_policy_init_v1(&raw)
		raw.keep_last = C.uint64_t(selected.keepLast)
		raw.collect_at = C.uint64_t(selected.collectAt)
		var output C.PotluckDbCollectionReportV1
		C.potluckdb_collection_report_init_v1(&output)
		cError := newError()
		if err := check(
			C.potluckdb_remote_collect_v1(
				handle, &raw, &output, C.uint32_t(unsafe.Sizeof(output)), &cError,
			),
			&cError,
		); err != nil {
			return err
		}
		report = collectionReport(&output)
		return nil
	})
	return report, err
}

// RecoverAbandonedCollection completes an interrupted collection.
func (r *Remote) RecoverAbandonedCollection() (*CollectionReport, error) {
	var report *CollectionReport
	err := r.state.withHandle(func(handle C.uint64_t) error {
		var output C.PotluckDbCollectionReportV1
		C.potluckdb_collection_report_init_v1(&output)
		cError := newError()
		if err := check(
			C.potluckdb_remote_recover_abandoned_collection_v1(
				handle, &output, C.uint32_t(unsafe.Sizeof(output)), &cError,
			),
			&cError,
		); err != nil {
			return err
		}
		report = collectionReport(&output)
		return nil
	})
	return report, err
}

func collectionReport(output *C.PotluckDbCollectionReportV1) *CollectionReport {
	report := &CollectionReport{
		Epoch:   uint64(output.epoch),
		Marked:  uint64(output.marked),
		Swept:   uint64(output.swept),
		Aborted: output.aborted != 0,
	}
	if output.has_floor_before != 0 {
		report.FloorBefore = readRevision(output.floor_before)
	}
	if output.has_floor_after != 0 {
		report.FloorAfter = readRevision(output.floor_after)
	}
	return report
}

// Close ends the Remote; Status stays observable afterward.
func (r *Remote) Close() error {
	return r.state.close(func(handle C.uint64_t) error {
		cError := newError()
		return check(C.potluckdb_remote_close_v1(handle, &cError), &cError)
	})
}

// Transaction owns one explicit transaction inside its callback.
type Transaction struct {
	state *handleState
}

func runTransaction(
	handle C.uint64_t, owner *handleState, body func(*Transaction) error,
) error {
	state, err := newHandleState(ownerTransaction, handle, false, owner)
	if err != nil {
		return err
	}
	transaction := &Transaction{state: state}
	if err := body(transaction); err != nil {
		transaction.rollbackQuietly()
		return err
	}
	return transaction.commit()
}

// Query reads rows inside the transaction.
func (t *Transaction) Query(sql string, bindings Bindings) ([][]any, error) {
	return runQuery(t.state, sql, bindings, func(
		handle C.uint64_t,
		slice C.PotluckDbSliceV1,
		raw *C.PotluckDbBindingsV1,
		rows *C.PotluckDbRowsHandleV1,
		cError *C.PotluckDbErrorV1,
	) C.int32_t {
		return C.potluckdb_transaction_query_v1(handle, slice, raw, rows, cError)
	})
}

// Execute writes inside the transaction.
func (t *Transaction) Execute(sql string, bindings Bindings) (*ExecuteResult, error) {
	return runExecute(t.state, sql, bindings, func(
		handle C.uint64_t,
		slice C.PotluckDbSliceV1,
		raw *C.PotluckDbBindingsV1,
		output *C.PotluckDbExecuteResultV1,
		size C.uint32_t,
		cError *C.PotluckDbErrorV1,
	) C.int32_t {
		return C.potluckdb_transaction_execute_v1(
			handle, slice, raw, output, size, cError,
		)
	})
}

func (t *Transaction) commit() error {
	if err := t.state.withHandle(func(handle C.uint64_t) error {
		cError := newError()
		return check(C.potluckdb_transaction_commit_v1(handle, &cError), &cError)
	}); err != nil {
		t.rollbackQuietly()
		return err
	}
	return t.state.release(func(handle C.uint64_t) error {
		cError := newError()
		return check(C.potluckdb_transaction_release_v1(handle, &cError), &cError)
	})
}

func (t *Transaction) rollbackQuietly() {
	// Rollback failures must not mask the original transaction failure.
	_ = t.state.withHandle(func(handle C.uint64_t) error {
		cError := newError()
		return check(C.potluckdb_transaction_rollback_v1(handle, &cError), &cError)
	})
	_ = t.state.release(func(handle C.uint64_t) error {
		cError := newError()
		return check(C.potluckdb_transaction_release_v1(handle, &cError), &cError)
	})
}

type queryCall func(
	C.uint64_t,
	C.PotluckDbSliceV1,
	*C.PotluckDbBindingsV1,
	*C.PotluckDbRowsHandleV1,
	*C.PotluckDbErrorV1,
) C.int32_t

type executeCall func(
	C.uint64_t,
	C.PotluckDbSliceV1,
	*C.PotluckDbBindingsV1,
	*C.PotluckDbExecuteResultV1,
	C.uint32_t,
	*C.PotluckDbErrorV1,
) C.int32_t

func runQuery(
	state *handleState, sql string, bindings Bindings, call queryCall,
) ([][]any, error) {
	var rows [][]any
	err := state.withHandle(func(handle C.uint64_t) error {
		return withUtf8(sql, func(slice C.PotluckDbSliceV1) error {
			return buildBindings(bindings, func(raw *C.PotluckDbBindingsV1) error {
				var output C.PotluckDbRowsHandleV1
				cError := newError()
				if err := check(
					call(handle, slice, raw, &output, &cError), &cError,
				); err != nil {
					return err
				}
				copied, err := copyRows(output)
				if err != nil {
					return err
				}
				rows = copied
				return nil
			})
		})
	})
	return rows, err
}

func runExecute(
	state *handleState, sql string, bindings Bindings, call executeCall,
) (*ExecuteResult, error) {
	var result *ExecuteResult
	err := state.withHandle(func(handle C.uint64_t) error {
		return withUtf8(sql, func(slice C.PotluckDbSliceV1) error {
			return buildBindings(bindings, func(raw *C.PotluckDbBindingsV1) error {
				var output C.PotluckDbExecuteResultV1
				C.potluckdb_execute_result_init_v1(&output)
				cError := newError()
				if err := check(
					call(
						handle,
						slice,
						raw,
						&output,
						C.uint32_t(unsafe.Sizeof(output)),
						&cError,
					),
					&cError,
				); err != nil {
					return err
				}
				result = &ExecuteResult{RowsAffected: int64(output.rows_affected)}
				if output.has_last_insert_id != 0 {
					last := int64(output.last_insert_id)
					result.LastInsertID = &last
				}
				return nil
			})
		})
	})
	return result, err
}

// Connection is one explicit connection to a Database.
type Connection struct {
	state *handleState
}

// Query reads rows through the connection.
func (c *Connection) Query(sql string, bindings Bindings) ([][]any, error) {
	return runQuery(c.state, sql, bindings, func(
		handle C.uint64_t,
		slice C.PotluckDbSliceV1,
		raw *C.PotluckDbBindingsV1,
		rows *C.PotluckDbRowsHandleV1,
		cError *C.PotluckDbErrorV1,
	) C.int32_t {
		return C.potluckdb_connection_query_v1(handle, slice, raw, rows, cError)
	})
}

// Execute writes through the connection.
func (c *Connection) Execute(sql string, bindings Bindings) (*ExecuteResult, error) {
	return runExecute(c.state, sql, bindings, func(
		handle C.uint64_t,
		slice C.PotluckDbSliceV1,
		raw *C.PotluckDbBindingsV1,
		output *C.PotluckDbExecuteResultV1,
		size C.uint32_t,
		cError *C.PotluckDbErrorV1,
	) C.int32_t {
		return C.potluckdb_connection_execute_v1(
			handle, slice, raw, output, size, cError,
		)
	})
}

// Transaction runs body inside one immediate transaction.
func (c *Connection) Transaction(body func(*Transaction) error) error {
	return c.TransactionWith(Immediate, body)
}

// TransactionWith runs body inside one transaction of the given mode.
func (c *Connection) TransactionWith(
	mode TransactionMode, body func(*Transaction) error,
) error {
	var transactionHandle C.PotluckDbTransactionHandleV1
	if err := c.state.withHandle(func(handle C.uint64_t) error {
		cError := newError()
		return check(
			C.potluckdb_connection_transaction_begin_v1(
				handle, C.uint32_t(mode), &transactionHandle, &cError,
			),
			&cError,
		)
	}); err != nil {
		return err
	}
	return runTransaction(transactionHandle, c.state, body)
}

// Checkpoint runs one manual WAL checkpoint.
func (c *Connection) Checkpoint() (*CheckpointResult, error) {
	var result *CheckpointResult
	err := c.state.withHandle(func(handle C.uint64_t) error {
		var output C.PotluckDbCheckpointResultV1
		C.potluckdb_checkpoint_result_init_v1(&output)
		cError := newError()
		if err := check(
			C.potluckdb_connection_checkpoint_v1(
				handle, &output, C.uint32_t(unsafe.Sizeof(output)), &cError,
			),
			&cError,
		); err != nil {
			return err
		}
		result = &CheckpointResult{
			LogFrames:          int32(output.log_frames),
			CheckpointedFrames: int32(output.checkpointed_frames),
		}
		return nil
	})
	return result, err
}

// Status reports the Connection observational status.
func (c *Connection) Status() (*ConnectionStatus, error) {
	var status *ConnectionStatus
	err := c.state.withHandle(func(handle C.uint64_t) error {
		var raw C.PotluckDbConnectionStatusV1
		C.potluckdb_connection_status_init_v1(&raw)
		cError := newError()
		if err := check(
			C.potluckdb_connection_status_v1(
				handle, &raw, C.uint32_t(unsafe.Sizeof(raw)), &cError,
			),
			&cError,
		); err != nil {
			return err
		}
		status = &ConnectionStatus{
			TransactionActive: raw.transaction_active != 0,
			Closed:            raw.closed != 0,
		}
		if raw.has_checkpoint != 0 {
			status.Checkpoint = &CheckpointStatus{
				SoftLimitBytes:           uint64(raw.checkpoint.soft_limit_bytes),
				ForcedLimitBytes:         uint64(raw.checkpoint.forced_checkpoint_threshold_bytes),
				WalBytes:                 uint64(raw.checkpoint.wal_bytes),
				Due:                      raw.checkpoint.due != 0,
				Scheduled:                raw.checkpoint.scheduled != 0,
				SoftScheduled:            uint64(raw.checkpoint.soft_scheduled),
				ForcedCheckpointAttempts: uint64(raw.checkpoint.forced_checkpoint_attempts),
				Manual:                   uint64(raw.checkpoint.manual),
				Completed:                uint64(raw.checkpoint.completed),
				Failures:                 uint64(raw.checkpoint.failures),
				StaleCommands:            uint64(raw.checkpoint.stale_commands),
				PagesVerified:            uint64(raw.checkpoint.pages_verified),
				WalBytesReclaimed:        uint64(raw.checkpoint.wal_bytes_reclaimed),
			}
		}
		if raw.has_prefetch != 0 {
			status.Prefetch = &PrefetchStatus{
				BackgroundState:       backgroundStateName(raw.prefetch.background_state),
				Limit:                 uint64(raw.prefetch.limit),
				Active:                uint64(raw.prefetch.active),
				Queued:                uint64(raw.prefetch.queued),
				ReadyPages:            uint64(raw.prefetch.ready_pages),
				ReadyBytes:            uint64(raw.prefetch.ready_bytes),
				BackgroundTargetBytes: uint64(raw.prefetch.background_target_bytes),
				PolicyFailures:        uint64(raw.prefetch.policy_failures),
			}
		}
		return nil
	})
	return status, err
}

// Close ends the Connection; Status stays observable afterward.
func (c *Connection) Close() error {
	return c.state.close(func(handle C.uint64_t) error {
		cError := newError()
		return check(C.potluckdb_connection_close_v1(handle, &cError), &cError)
	})
}

// Snapshot is one pinned read-only revision session.
type Snapshot struct {
	state *handleState
}

// Query reads rows from the pinned revision.
func (s *Snapshot) Query(sql string, bindings Bindings) ([][]any, error) {
	return runQuery(s.state, sql, bindings, func(
		handle C.uint64_t,
		slice C.PotluckDbSliceV1,
		raw *C.PotluckDbBindingsV1,
		rows *C.PotluckDbRowsHandleV1,
		cError *C.PotluckDbErrorV1,
	) C.int32_t {
		return C.potluckdb_snapshot_query_v1(handle, slice, raw, rows, cError)
	})
}

// Status reports the Snapshot observational status.
func (s *Snapshot) Status() (*SnapshotStatus, error) {
	var status *SnapshotStatus
	err := s.state.withHandle(func(handle C.uint64_t) error {
		var raw C.PotluckDbSnapshotStatusV1
		C.potluckdb_snapshot_status_init_v1(&raw)
		cError := newError()
		if err := check(
			C.potluckdb_snapshot_status_v1(
				handle, &raw, C.uint32_t(unsafe.Sizeof(raw)), &cError,
			),
			&cError,
		); err != nil {
			return err
		}
		status = &SnapshotStatus{
			Revision:   readRevision(raw.revision),
			LocalBytes: uint64(raw.local_bytes),
			Closed:     raw.closed != 0,
		}
		return nil
	})
	return status, err
}

// Close ends the Snapshot; Status stays observable afterward.
func (s *Snapshot) Close() error {
	return s.state.close(func(handle C.uint64_t) error {
		cError := newError()
		return check(C.potluckdb_snapshot_close_v1(handle, &cError), &cError)
	})
}

// Database is one writable session.
type Database struct {
	state *handleState
}

// Open opens the current writable database; nil remote opens unbound and nil
// options selects the defaults.
func Open(path string, remote *Remote, options *OpenOptions) (*Database, error) {
	selected, err := options.normalized()
	if err != nil {
		return nil, err
	}
	open := func(remoteHandle C.uint64_t, retained any) (*Database, error) {
		var handle C.PotluckDbDatabaseHandleV1
		err := withUtf8(path, func(slice C.PotluckDbSliceV1) error {
			return withOpenOptions(selected, func(raw *C.PotluckDbOpenOptionsV1) error {
				cError := newError()
				return check(
					C.potluckdb_open_v1(slice, remoteHandle, raw, &handle, &cError),
					&cError,
				)
			})
		})
		if err != nil {
			return nil, err
		}
		state, err := newHandleState(ownerDatabase, handle, true, retained)
		if err != nil {
			return nil, err
		}
		return &Database{state: state}, nil
	}
	if remote == nil {
		return open(0, nil)
	}
	var database *Database
	err = remote.state.withHandle(func(remoteHandle C.uint64_t) error {
		opened, openErr := open(remoteHandle, remote.state)
		if openErr != nil {
			return openErr
		}
		database = opened
		return nil
	})
	return database, err
}

// OpenRevision opens one pinned read-only revision; nil options selects the
// defaults.
func OpenRevision(
	path string, revision *Revision, remote *Remote, options *OpenOptions,
) (*Snapshot, error) {
	selected, err := options.normalized()
	if err != nil {
		return nil, err
	}
	var snapshot *Snapshot
	err = remote.state.withHandle(func(remoteHandle C.uint64_t) error {
		raw := writeRevision(revision)
		var handle C.PotluckDbSnapshotHandleV1
		openErr := withUtf8(path, func(slice C.PotluckDbSliceV1) error {
			return withOpenOptions(selected, func(rawOptions *C.PotluckDbOpenOptionsV1) error {
				cError := newError()
				return check(
					C.potluckdb_open_revision_v1(
						slice, &raw, remoteHandle, rawOptions, &handle, &cError,
					),
					&cError,
				)
			})
		})
		if openErr != nil {
			return openErr
		}
		state, stateErr := newHandleState(ownerSnapshot, handle, true, remote.state)
		if stateErr != nil {
			return stateErr
		}
		snapshot = &Snapshot{state: state}
		return nil
	})
	return snapshot, err
}

func withOpenOptions(
	options *OpenOptions, body func(*C.PotluckDbOpenOptionsV1) error,
) error {
	var raw C.PotluckDbOpenOptionsV1
	C.potluckdb_open_options_init_v1(&raw)
	raw.local_budget_bytes = C.uint64_t(options.LocalBudgetBytes)
	raw.minimum_local_budget_bytes = C.uint64_t(options.MinimumLocalBudgetBytes)
	raw.local_compaction_bytes = C.uint64_t(options.LocalCompactionBytes)
	C.potluckdb_sync_policy_init_v1(&raw.sync_policy)
	fillSyncPolicy(&raw.sync_policy, options.SyncPolicy)
	C.potluckdb_connection_options_init_v1(&raw.connection)
	fillConnectionOptions(&raw.connection, options.Connection)
	if options.EncryptionKey != nil {
		raw.flags |= 1
		C.memcpy(
			unsafe.Pointer(&raw.encryption_key[0]),
			unsafe.Pointer(&options.EncryptionKey[0]),
			32,
		)
	}
	if options.ClientID != "" {
		clientID := []byte(options.ClientID)
		raw.flags |= 2
		raw.client_id = C.PotluckDbSliceV1{
			data:   (*C.uint8_t)(unsafe.Pointer(&clientID[0])),
			length: C.uint64_t(len(clientID)),
		}
		err := body(&raw)
		runtime.KeepAlive(clientID)
		return err
	}
	return body(&raw)
}

func fillSyncPolicy(raw *C.PotluckDbSyncPolicyV1, policy *SyncPolicy) {
	if policy.automatic {
		raw.mode = 1
	} else {
		raw.mode = 2
	}
	raw.outbound_debounce_ms = C.uint64_t(policy.options.OutboundDebounce / time.Millisecond)
	raw.active_pull_interval_ms = C.uint64_t(policy.options.ActivePullInterval / time.Millisecond)
	raw.idle_pull_interval_ms = C.uint64_t(policy.options.IdlePullInterval / time.Millisecond)
	raw.retry_initial_ms = C.uint64_t(policy.options.RetryInitial / time.Millisecond)
	raw.retry_max_ms = C.uint64_t(policy.options.RetryMaximum / time.Millisecond)
}

func fillConnectionOptions(
	raw *C.PotluckDbConnectionOptionsV1, options *ConnectionOptions,
) {
	raw.timeout_ms = C.uint64_t(options.Timeout / time.Millisecond)
	raw.maximum_rows = C.uint64_t(options.MaximumRows)
	if options.Prefetch.enabled {
		raw.prefetch_mode = 1
	} else {
		raw.prefetch_mode = 2
	}
	raw.prefetch_background_target_bytes =
		C.uint64_t(options.Prefetch.backgroundTargetBytes)
	raw.prefetch_memory_budget_bytes = C.uint64_t(options.Prefetch.memoryBudgetBytes)
	if options.Checkpoint.enabled {
		raw.checkpoint_mode = 1
	} else {
		raw.checkpoint_mode = 2
	}
	raw.checkpoint_soft_limit_bytes = C.uint64_t(options.Checkpoint.softLimitBytes)
	raw.checkpoint_forced_limit_bytes = C.uint64_t(options.Checkpoint.forcedLimitBytes)
}

// Query reads rows through the default connection.
func (d *Database) Query(sql string, bindings Bindings) ([][]any, error) {
	return runQuery(d.state, sql, bindings, func(
		handle C.uint64_t,
		slice C.PotluckDbSliceV1,
		raw *C.PotluckDbBindingsV1,
		rows *C.PotluckDbRowsHandleV1,
		cError *C.PotluckDbErrorV1,
	) C.int32_t {
		return C.potluckdb_database_query_v1(handle, slice, raw, rows, cError)
	})
}

// Execute writes through the default connection.
func (d *Database) Execute(sql string, bindings Bindings) (*ExecuteResult, error) {
	return runExecute(d.state, sql, bindings, func(
		handle C.uint64_t,
		slice C.PotluckDbSliceV1,
		raw *C.PotluckDbBindingsV1,
		output *C.PotluckDbExecuteResultV1,
		size C.uint32_t,
		cError *C.PotluckDbErrorV1,
	) C.int32_t {
		return C.potluckdb_database_execute_v1(
			handle, slice, raw, output, size, cError,
		)
	})
}

// Transaction runs body inside one immediate transaction.
func (d *Database) Transaction(body func(*Transaction) error) error {
	return d.TransactionWith(Immediate, body)
}

// TransactionWith runs body inside one transaction of the given mode.
func (d *Database) TransactionWith(
	mode TransactionMode, body func(*Transaction) error,
) error {
	var transactionHandle C.PotluckDbTransactionHandleV1
	if err := d.state.withHandle(func(handle C.uint64_t) error {
		cError := newError()
		return check(
			C.potluckdb_database_transaction_begin_v1(
				handle, C.uint32_t(mode), &transactionHandle, &cError,
			),
			&cError,
		)
	}); err != nil {
		return err
	}
	return runTransaction(transactionHandle, d.state, body)
}

// Connect opens one explicit connection; nil options selects the defaults.
func (d *Database) Connect(options *ConnectionOptions) (*Connection, error) {
	selected, err := options.normalized()
	if err != nil {
		return nil, err
	}
	var connection *Connection
	err = d.state.withHandle(func(handle C.uint64_t) error {
		var raw C.PotluckDbConnectionOptionsV1
		C.potluckdb_connection_options_init_v1(&raw)
		fillConnectionOptions(&raw, selected)
		var connectionHandle C.PotluckDbConnectionHandleV1
		cError := newError()
		if err := check(
			C.potluckdb_database_connect_v1(
				handle, &raw, &connectionHandle, &cError,
			),
			&cError,
		); err != nil {
			return err
		}
		state, stateErr := newHandleState(
			ownerConnection, connectionHandle, true, d.state,
		)
		if stateErr != nil {
			return stateErr
		}
		connection = &Connection{state: state}
		return nil
	})
	return connection, err
}

// Sync forces one complete bidirectional synchronization cycle.
func (d *Database) Sync() (SyncResult, error) {
	var result SyncResult
	err := d.state.withHandle(func(handle C.uint64_t) error {
		var output C.PotluckDbSyncResultV1
		C.potluckdb_sync_result_init_v1(&output)
		cError := newError()
		if err := check(
			C.potluckdb_database_sync_v1(
				handle, &output, C.uint32_t(unsafe.Sizeof(output)), &cError,
			),
			&cError,
		); err != nil {
			return err
		}
		defer C.potluckdb_sync_result_release_v1(&output)
		switch output.kind {
		case 1:
			value := UpToDate{}
			if output.has_revision != 0 {
				value.Revision = readRevision(output.revision)
			}
			result = value
		case 2:
			result = Pushed{Revision: readRevision(output.revision)}
		case 3:
			replayed, err := copyStrings(output.replayed)
			if err != nil {
				return err
			}
			quarantined, err := copyStrings(output.quarantined)
			if err != nil {
				return err
			}
			result = Pulled{
				Revision:    readRevision(output.revision),
				Replayed:    replayed,
				Quarantined: quarantined,
			}
		case 4:
			value := Conflict{Reason: bufferString(output.reason)}
			if output.has_revision != 0 {
				value.RemoteRevision = readRevision(output.revision)
			}
			result = value
		default:
			return internalFailure(
				"potluckdb returned an unknown synchronization result",
			)
		}
		return nil
	})
	return result, err
}

// ResolveConflict resolves the current conflict through authenticated rebase.
func (d *Database) ResolveConflict() (*ConflictResolution, error) {
	var resolution *ConflictResolution
	err := d.state.withHandle(func(handle C.uint64_t) error {
		var output C.PotluckDbConflictResolutionV1
		C.potluckdb_conflict_resolution_init_v1(&output)
		cError := newError()
		if err := check(
			C.potluckdb_database_resolve_conflict_v1(
				handle, &output, C.uint32_t(unsafe.Sizeof(output)), &cError,
			),
			&cError,
		); err != nil {
			return err
		}
		defer C.potluckdb_conflict_resolution_release_v1(&output)
		alreadyPublished, err := copyStrings(output.already_published)
		if err != nil {
			return err
		}
		replayed, err := copyStrings(output.replayed)
		if err != nil {
			return err
		}
		quarantined, err := copyStrings(output.quarantined)
		if err != nil {
			return err
		}
		pending, err := copyStrings(output.pending)
		if err != nil {
			return err
		}
		resolution = &ConflictResolution{
			Revision:         readRevision(output.revision),
			AlreadyPublished: alreadyPublished,
			Replayed:         replayed,
			Quarantined:      quarantined,
			Pending:          pending,
		}
		return nil
	})
	return resolution, err
}

// SetSyncPolicy switches synchronization policy and reports the new status.
func (d *Database) SetSyncPolicy(policy *SyncPolicy) (*DatabaseStatus, error) {
	var status *DatabaseStatus
	err := d.state.withHandle(func(handle C.uint64_t) error {
		var raw C.PotluckDbSyncPolicyV1
		C.potluckdb_sync_policy_init_v1(&raw)
		fillSyncPolicy(&raw, policy)
		var output C.PotluckDbDatabaseStatusV1
		C.potluckdb_database_status_init_v1(&output)
		cError := newError()
		if err := check(
			C.potluckdb_database_set_sync_policy_v1(
				handle, &raw, &output, C.uint32_t(unsafe.Sizeof(output)), &cError,
			),
			&cError,
		); err != nil {
			return err
		}
		status = databaseStatus(&output)
		return nil
	})
	return status, err
}

// ClearQuarantineRecords drops quarantined transaction records.
func (d *Database) ClearQuarantineRecords() (uint64, error) {
	var count uint64
	err := d.state.withHandle(func(handle C.uint64_t) error {
		var output C.uint64_t
		cError := newError()
		if err := check(
			C.potluckdb_database_clear_quarantine_records_v1(
				handle, &output, &cError,
			),
			&cError,
		); err != nil {
			return err
		}
		count = uint64(output)
		return nil
	})
	return count, err
}

// Status reports the Database observational status.
func (d *Database) Status() (*DatabaseStatus, error) {
	var status *DatabaseStatus
	err := d.state.withHandle(func(handle C.uint64_t) error {
		var output C.PotluckDbDatabaseStatusV1
		C.potluckdb_database_status_init_v1(&output)
		cError := newError()
		if err := check(
			C.potluckdb_database_status_v1(
				handle, &output, C.uint32_t(unsafe.Sizeof(output)), &cError,
			),
			&cError,
		); err != nil {
			return err
		}
		status = databaseStatus(&output)
		return nil
	})
	return status, err
}

func databaseStatus(output *C.PotluckDbDatabaseStatusV1) *DatabaseStatus {
	status := &DatabaseStatus{
		PendingWrites:    uint64(output.pending_writes),
		LocalBytes:       uint64(output.local_bytes),
		LocalBudgetBytes: uint64(output.local_budget_bytes),
		ConnectionCount:  uint64(output.connection_count),
		SyncState:        syncStateName(output.sync_state),
		Closed:           output.closed != 0,
	}
	if output.has_revision != 0 {
		status.Revision = readRevision(output.revision)
	}
	if output.has_next_sync != 0 {
		next := time.Duration(output.next_sync_in_ms) * time.Millisecond
		status.NextSyncIn = &next
	}
	if output.has_conflict_revision != 0 {
		status.ConflictRemoteRevision = readRevision(output.conflict_revision)
	}
	return status
}

// Close ends the Database; Status stays observable afterward.
func (d *Database) Close() error {
	return d.state.close(func(handle C.uint64_t) error {
		cError := newError()
		return check(C.potluckdb_database_close_v1(handle, &cError), &cError)
	})
}

// ReencryptOptions selects optional source key and exact revision.
type ReencryptOptions struct {
	SourceKey []byte
	Revision  *Revision
}

// Reencrypt republishes source into target under a new 32-byte key.
func Reencrypt(
	source, target *Remote, newKey []byte, options *ReencryptOptions,
) (*KeyRotationReport, error) {
	if len(newKey) != 32 {
		return nil, invalidArgument(
			"encryption key must contain exactly 32 bytes",
			"Supply a 32-byte key and retry.",
		)
	}
	if options != nil && options.SourceKey != nil && len(options.SourceKey) != 32 {
		return nil, invalidArgument(
			"source encryption key must contain exactly 32 bytes",
			"Supply a 32-byte key and retry.",
		)
	}
	var raw C.PotluckDbReencryptOptionsV1
	C.potluckdb_reencrypt_options_init_v1(&raw)
	C.memcpy(unsafe.Pointer(&raw.new_key[0]), unsafe.Pointer(&newKey[0]), 32)
	if options != nil && options.SourceKey != nil {
		raw.flags |= 1
		C.memcpy(
			unsafe.Pointer(&raw.source_key[0]),
			unsafe.Pointer(&options.SourceKey[0]),
			32,
		)
	}
	if options != nil && options.Revision != nil {
		raw.flags |= 2
		raw.revision = writeRevision(options.Revision)
	}
	var report *KeyRotationReport
	err := withPair(
		source.state,
		target.state,
		func(left, right C.uint64_t) error {
			var output C.PotluckDbKeyRotationReportV1
			C.potluckdb_key_rotation_report_init_v1(&output)
			cError := newError()
			if err := check(
				C.potluckdb_admin_reencrypt_v1(
					left,
					right,
					&raw,
					&output,
					C.uint32_t(unsafe.Sizeof(output)),
					&cError,
				),
				&cError,
			); err != nil {
				return err
			}
			report = &KeyRotationReport{
				SourceRevision: readRevision(output.source_revision),
				TargetRevision: readRevision(output.target_revision),
				TargetKeyID:    readRevision(output.target_key_id),
				PieceCount:     uint64(output.piece_count),
				Published:      output.published != 0,
			}
			if output.has_source_key_id != 0 {
				report.SourceKeyID = readRevision(output.source_key_id)
			}
			return nil
		},
	)
	return report, err
}
