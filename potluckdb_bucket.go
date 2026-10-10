package potluckdb

/*
#cgo CFLAGS: -I${SRCDIR}

#include <stdlib.h>

#include "potluckdb_bucket_bridge.h"
*/
import "C"

import (
	"fmt"
	"reflect"
	"runtime"
	"runtime/cgo"
	"strings"
	"sync"
	"unsafe"
)

// The Seam-4 error codes a bucket answers in. They are the Runtime's own
// (POTLUCKDB_RUNTIME_ERROR_* in native/interface/include/potluckdb_runtime.h),
// because a bucket is called by the Runtime across Seam 4 rather than by
// an application: it never sees the string codes the facade reports.
//
// [Seam 4 B5] fixes the three a store reports -- absent is Missing,
// immutable disagreement is Conflict, and a transport or permission
// failure is Unavailable. BucketErrorContractViolation is for a bucket
// that answered outside its own interface, and
// BucketErrorInternalFailure is what an unclassified failure becomes.
const (
	BucketErrorInternalFailure   int32 = 1
	BucketErrorMissing           int32 = 2
	BucketErrorConflict          int32 = 6
	BucketErrorUnavailable       int32 = 7
	BucketErrorContractViolation int32 = 18
	BucketErrorBucketNotFound    int32 = 24
)

// bucketErrorCodes is every code a host may report: the Runtime
// Interface's codes except INVALID_ARGUMENT (25), which only the Runtime
// raises. A bucket answering outside this set is a contract violation
// rather than a failure the Runtime can interpret.
var bucketErrorCodes = map[int32]struct{}{
	1: {}, 2: {}, 3: {}, 4: {}, 5: {}, 6: {}, 7: {}, 8: {}, 9: {}, 10: {},
	11: {}, 12: {}, 13: {}, 14: {}, 15: {}, 16: {}, 17: {}, 18: {}, 19: {},
	20: {}, 21: {}, 22: {}, 23: {}, 24: {},
}

// BucketError is a typed Seam-4 failure. Any error carrying an int32
// field named Code holding one of the Runtime's error codes is read the
// same way, so an adapter package need not import this one to answer in
// the seam's vocabulary.
//
// Integrator API: for code that supplies its own storage transport. Application
// code reaches a bucket through FromS3, FromGCS and FromAzure.
type BucketError struct {
	Code    int32
	Message string
}

func (e *BucketError) Error() string { return e.Message }

// BucketCounterSnapshot is the observational Seam-4 counter snapshot.
// [Seam 4 B12] separates the logical operations the seam names from the
// provider API calls they cause.
//
// Integrator API: for code that supplies its own storage transport. Application
// code reaches a bucket through FromS3, FromGCS and FromAzure.
type BucketCounterSnapshot struct {
	Gets                    uint64
	Puts                    uint64
	IdempotentPuts          uint64
	Exists                  uint64
	Lists                   uint64
	Deletes                 uint64
	CellGets                uint64
	CellSwaps               uint64
	CellSwapWins            uint64
	TransportGets           uint64
	TransportPuts           uint64
	TransportHeads          uint64
	TransportLists          uint64
	TransportDeletes        uint64
	TransportBytesReceived  uint64
	TransportBytesSubmitted uint64
}

// bucketCounterFields is the exact member set a Counters method must
// return, in the order the Runtime's snapshot declares them.
var bucketCounterFields = [16]string{
	"Gets",
	"Puts",
	"IdempotentPuts",
	"Exists",
	"Lists",
	"Deletes",
	"CellGets",
	"CellSwaps",
	"CellSwapWins",
	"TransportGets",
	"TransportPuts",
	"TransportHeads",
	"TransportLists",
	"TransportDeletes",
	"TransportBytesReceived",
	"TransportBytesSubmitted",
}

// BucketStorage is one Seam-4 bucket: immutable objects plus the head and
// floor compare-and-swap cells.
//
// An implementor must also carry a Counters method returning a struct
// with the sixteen uint64 fields BucketCounterSnapshot declares. That one
// member is matched structurally rather than by type, so an adapter in
// its own package -- one that must not drag this cgo facade into a build
// that only needs the adapter -- satisfies it by declaring its own
// snapshot type with the same fields. Everything else is matched by the
// ordinary Go rules.
//
// [Seam 4 B10] and [B11] let immutable get and put calls overlap on one
// instance, so every method must be safe for concurrent use. Mutable cell
// calls stay on the exchange caller.
//
// Integrator API: for code that supplies its own storage transport. Application
// code reaches a bucket through FromS3, FromGCS and FromAzure.
type BucketStorage interface {
	Get(key string) ([]byte, error)
	Put(key string, data []byte) error
	Exists(key string) (bool, error)
	List(prefix string) ([]string, error)
	Delete(key string) error
	HeadGet() ([]byte, error)
	HeadSwap(expected, replacement []byte) (bool, error)
	FloorGet() ([]byte, error)
	FloorSwap(expected, replacement []byte) (bool, error)
}

// CellPairBucketStorage is the optional [Seam 4 B13] capability: one
// finite concurrent head-and-floor observation. A nil value is an absent
// cell. A bucket that does not implement it gets the ordinary serial
// HeadGet then FloorGet fallback.
//
// Integrator API: for code that supplies its own storage transport. Application
// code reaches a bucket through FromS3, FromGCS and FromAzure.
type CellPairBucketStorage interface {
	BucketStorage
	CellPairGet() (head, floor []byte, err error)
}

// bucketCounters reads a structurally matched Counters method.
type bucketCounters struct {
	method  reflect.Value
	indexes [16]int
}

func newBucketCounters(storage BucketStorage) (*bucketCounters, error) {
	method := reflect.ValueOf(storage).MethodByName("Counters")
	if !method.IsValid() {
		return nil, invalidArgument(
			"bucket storage must carry a Counters method",
			"Add Counters() returning a snapshot with the sixteen Seam-4 counters.",
		)
	}
	signature := method.Type()
	if signature.NumIn() != 0 || signature.NumOut() != 1 {
		return nil, invalidArgument(
			"bucket storage Counters must take no arguments and return one snapshot",
			"Match Counters() BucketCounterSnapshot and retry.",
		)
	}
	snapshot := signature.Out(0)
	if snapshot.Kind() != reflect.Struct {
		return nil, invalidArgument(
			"bucket storage Counters must return a struct",
			"Return a snapshot with the sixteen Seam-4 counters.",
		)
	}
	counters := &bucketCounters{method: method}
	for position, name := range bucketCounterFields {
		field, found := snapshot.FieldByName(name)
		if !found || field.Type.Kind() != reflect.Uint64 || len(field.Index) != 1 {
			return nil, invalidArgument(
				fmt.Sprintf(
					"bucket storage counters are missing the uint64 field %s", name,
				),
				"Declare all sixteen Seam-4 counters as uint64 fields and retry.",
			)
		}
		counters.indexes[position] = field.Index[0]
	}
	return counters, nil
}

func (c *bucketCounters) snapshot() BucketCounterSnapshot {
	value := c.method.Call(nil)[0]
	var result BucketCounterSnapshot
	target := reflect.ValueOf(&result).Elem()
	for position := range bucketCounterFields {
		target.Field(position).SetUint(value.Field(c.indexes[position]).Uint())
	}
	return result
}

// bucketMessageLimit bounds retained host-error message storage. A
// message stays readable for the adapter's whole life, which is what the
// Application ABI asks for when callbacks may overlap; a bucket that
// fails with a fresh message every time would otherwise grow that store
// without bound.
const bucketMessageLimit = 4096

var bucketMessageOverflow = "bucket adapter error message retention limit reached"

// bucketBridge owns one Go implementor and the C storage its adapter
// table points at.
//
// Lifetime: the Runtime copies the table but keeps the context pointer,
// so the C block and the cgo handle naming this bridge must outlive every
// callback. The bridge is released only when the StorageAdapter holding
// it becomes unreachable, which cannot happen while the Remote that
// retains it is alive.
type bucketBridge struct {
	storage  BucketStorage
	cellPair CellPairBucketStorage
	counters *bucketCounters

	block *C.PotluckDbGoBucketBridgeV1
	self  cgo.Handle

	// mutex guards messages only. It is never held across a call into
	// the Go implementor, which would serialize the overlapping gets and
	// puts B10 and B11 require.
	mutex    sync.Mutex
	messages map[string]unsafe.Pointer
}

func bucketBridgeFor(context unsafe.Pointer) (bridge *bucketBridge) {
	// A released handle makes Value panic, and a panic here would cross C
	// with no recover site above it. Treating that as a missing context
	// reports the contract violation the seam already has a code for.
	defer func() {
		if recover() != nil {
			bridge = nil
		}
	}()
	if context == nil {
		return nil
	}
	handle := cgo.Handle(*(*C.uintptr_t)(context))
	bridge, _ = handle.Value().(*bucketBridge)
	return bridge
}

// retain copies message into C storage that stays valid for the bridge's
// lifetime, so the Runtime may read it any time within the enclosing ABI
// call rather than only inside the callback.
func (b *bucketBridge) retain(message string) (unsafe.Pointer, int) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	if retained, found := b.messages[message]; found {
		return retained, len(message)
	}
	if len(b.messages) >= bucketMessageLimit {
		message = bucketMessageOverflow
		if retained, found := b.messages[message]; found {
			return retained, len(message)
		}
	}
	retained := C.CBytes([]byte(message))
	b.messages[message] = retained
	return retained, len(message)
}

// reclaimed reports whether the C adapter block has already been freed.
func (b *bucketBridge) reclaimed() bool {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	return b.block == nil
}

func (b *bucketBridge) release() {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	for _, retained := range b.messages {
		C.free(retained)
	}
	b.messages = nil
	if b.block != nil {
		C.potluckdb_go_bucket_bridge_free(b.block)
		b.block = nil
	}
	b.self.Delete()
}

// seamCode classifies one implementor failure.
//
// An error carrying an int32 Code field holding a Runtime error code is
// taken at its word, wherever in the chain it sits; an int32 Code outside
// that set is a contract violation, because a bucket that names a code
// the seam does not define has answered outside its interface. Anything
// else is an unclassified internal failure: guessing Unavailable would
// read as a transport problem the caller could retry, and guessing
// Missing would read as a satisfied absence.
func adapterErrorMessage(failure error) (message string) {
	message = fmt.Sprintf("storage adapter failed (%T)", failure)
	defer func() { _ = recover() }()
	return failure.Error()
}

func seamCode(failure error) (int32, string) {
	message := adapterErrorMessage(failure)
	pending := []error{failure}
	for len(pending) != 0 {
		current := pending[len(pending)-1]
		pending = pending[:len(pending)-1]
		if current == nil {
			continue
		}
		value := reflect.ValueOf(current)
		if value.Kind() == reflect.Pointer && !value.IsNil() {
			value = value.Elem()
		}
		if value.Kind() == reflect.Struct {
			field := value.FieldByName("Code")
			if field.IsValid() && field.Type() == reflect.TypeOf(int32(0)) {
				code := int32(field.Int())
				if _, known := bucketErrorCodes[code]; known {
					return code, message
				}
				return BucketErrorContractViolation, fmt.Sprintf(
					"Go bucket adapter returned the undefined error code %d: %s",
					code,
					message,
				)
			}
		}
		switch wrapped := current.(type) {
		case interface{ Unwrap() []error }:
			children := wrapped.Unwrap()
			for index := len(children) - 1; index >= 0; index-- {
				pending = append(pending, children[index])
			}
		case interface{ Unwrap() error }:
			pending = append(pending, wrapped.Unwrap())
		}
	}
	return BucketErrorInternalFailure, message
}

func resetHostError(failure *C.PotluckDbRuntimeHostErrorV1) {
	if failure == nil {
		return
	}
	failure.struct_size = C.uint32_t(unsafe.Sizeof(*failure))
	failure.abi_version = C.POTLUCKDB_RUNTIME_ABI_VERSION_V1
	failure.sqlite_primary_code = 0
	failure.sqlite_extended_code = 0
	failure.message.data = nil
	failure.message.length = 0
}

func (b *bucketBridge) fail(
	failure *C.PotluckDbRuntimeHostErrorV1, err error,
) C.int32_t {
	code, message := seamCode(err)
	// The runtime refuses a host message that is not UTF-8 and drops
	// the code with it, so the message is sanitized and the code kept.
	message = strings.ToValidUTF8(message, "�")
	retained, length := b.retain(message)
	if failure != nil {
		failure.struct_size = C.uint32_t(unsafe.Sizeof(*failure))
		failure.abi_version = C.POTLUCKDB_RUNTIME_ABI_VERSION_V1
		failure.sqlite_primary_code = 0
		failure.sqlite_extended_code = 0
		failure.message.data = (*C.uint8_t)(retained)
		failure.message.length = C.uint64_t(length)
	}
	return C.int32_t(code)
}

// guard runs one callback body. A panic must not unwind into C, where
// there is no Go stack to recover on, so it becomes an internal failure
// exactly as the Rust binding's catch_unwind does.
func (b *bucketBridge) guard(
	operation string,
	failure *C.PotluckDbRuntimeHostErrorV1,
	body func() C.int32_t,
) (code C.int32_t) {
	defer func() {
		if panicked := recover(); panicked != nil {
			code = b.fail(failure, &BucketError{
				Code: BucketErrorInternalFailure,
				Message: fmt.Sprintf(
					"Go bucket %s callback panicked: %v", operation, panicked,
				),
			})
		}
	}()
	return body()
}

func bucketString(data *C.uint8_t, length C.uint64_t) string {
	if length == 0 {
		return ""
	}
	return C.GoStringN((*C.char)(unsafe.Pointer(data)), C.int(length))
}

func bucketBytes(data *C.uint8_t, length C.uint64_t) []byte {
	if length == 0 {
		return []byte{}
	}
	return C.GoBytes(unsafe.Pointer(data), C.int(length))
}

func bucketEmit(
	emit C.PotluckDbRuntimeEmitBytesV1, emitContext unsafe.Pointer, value []byte,
) C.int32_t {
	if len(value) == 0 {
		return C.potluckdb_go_bucket_emit(emit, emitContext, nil, 0)
	}
	return C.potluckdb_go_bucket_emit(
		emit,
		emitContext,
		(*C.uint8_t)(unsafe.Pointer(&value[0])),
		C.uint64_t(len(value)),
	)
}

//export potluckdbGoBucketGet
func potluckdbGoBucketGet(
	context unsafe.Pointer,
	key *C.uint8_t,
	keyLength C.uint64_t,
	emit C.PotluckDbRuntimeEmitBytesV1,
	emitContext unsafe.Pointer,
	failure *C.PotluckDbRuntimeHostErrorV1,
) C.int32_t {
	resetHostError(failure)
	bridge := bucketBridgeFor(context)
	if bridge == nil {
		return C.int32_t(BucketErrorContractViolation)
	}
	return bridge.guard("get", failure, func() C.int32_t {
		value, err := bridge.storage.Get(bucketString(key, keyLength))
		if err != nil {
			return bridge.fail(failure, err)
		}
		return bucketEmit(emit, emitContext, value)
	})
}

//export potluckdbGoBucketPut
func potluckdbGoBucketPut(
	context unsafe.Pointer,
	key *C.uint8_t,
	keyLength C.uint64_t,
	data *C.uint8_t,
	dataLength C.uint64_t,
	failure *C.PotluckDbRuntimeHostErrorV1,
) C.int32_t {
	resetHostError(failure)
	bridge := bucketBridgeFor(context)
	if bridge == nil {
		return C.int32_t(BucketErrorContractViolation)
	}
	return bridge.guard("put", failure, func() C.int32_t {
		err := bridge.storage.Put(
			bucketString(key, keyLength), bucketBytes(data, dataLength),
		)
		if err != nil {
			return bridge.fail(failure, err)
		}
		return 0
	})
}

//export potluckdbGoBucketExists
func potluckdbGoBucketExists(
	context unsafe.Pointer,
	key *C.uint8_t,
	keyLength C.uint64_t,
	outExists *C.uint32_t,
	failure *C.PotluckDbRuntimeHostErrorV1,
) C.int32_t {
	resetHostError(failure)
	bridge := bucketBridgeFor(context)
	if bridge == nil {
		return C.int32_t(BucketErrorContractViolation)
	}
	return bridge.guard("exists", failure, func() C.int32_t {
		if outExists == nil {
			return bridge.fail(failure, &BucketError{
				Code:    BucketErrorContractViolation,
				Message: "bucket exists output is null",
			})
		}
		present, err := bridge.storage.Exists(bucketString(key, keyLength))
		if err != nil {
			return bridge.fail(failure, err)
		}
		if present {
			*outExists = 1
		} else {
			*outExists = 0
		}
		return 0
	})
}

//export potluckdbGoBucketList
func potluckdbGoBucketList(
	context unsafe.Pointer,
	prefix *C.uint8_t,
	prefixLength C.uint64_t,
	emit C.PotluckDbRuntimeEmitBytesV1,
	emitContext unsafe.Pointer,
	failure *C.PotluckDbRuntimeHostErrorV1,
) C.int32_t {
	resetHostError(failure)
	bridge := bucketBridgeFor(context)
	if bridge == nil {
		return C.int32_t(BucketErrorContractViolation)
	}
	return bridge.guard("list", failure, func() C.int32_t {
		keys, err := bridge.storage.List(bucketString(prefix, prefixLength))
		if err != nil {
			return bridge.fail(failure, err)
		}
		for _, name := range keys {
			if code := bucketEmit(emit, emitContext, []byte(name)); code != 0 {
				return code
			}
		}
		return 0
	})
}

//export potluckdbGoBucketDelete
func potluckdbGoBucketDelete(
	context unsafe.Pointer,
	key *C.uint8_t,
	keyLength C.uint64_t,
	failure *C.PotluckDbRuntimeHostErrorV1,
) C.int32_t {
	resetHostError(failure)
	bridge := bucketBridgeFor(context)
	if bridge == nil {
		return C.int32_t(BucketErrorContractViolation)
	}
	return bridge.guard("delete", failure, func() C.int32_t {
		if err := bridge.storage.Delete(bucketString(key, keyLength)); err != nil {
			return bridge.fail(failure, err)
		}
		return 0
	})
}

//export potluckdbGoBucketCellGet
func potluckdbGoBucketCellGet(
	context unsafe.Pointer,
	head C.int,
	emit C.PotluckDbRuntimeEmitBytesV1,
	emitContext unsafe.Pointer,
	failure *C.PotluckDbRuntimeHostErrorV1,
) C.int32_t {
	resetHostError(failure)
	bridge := bucketBridgeFor(context)
	if bridge == nil {
		return C.int32_t(BucketErrorContractViolation)
	}
	return bridge.guard("cell-get", failure, func() C.int32_t {
		var value []byte
		var err error
		if head != 0 {
			value, err = bridge.storage.HeadGet()
		} else {
			value, err = bridge.storage.FloorGet()
		}
		if err != nil {
			return bridge.fail(failure, err)
		}
		return bucketEmit(emit, emitContext, value)
	})
}

//export potluckdbGoBucketCellSwap
func potluckdbGoBucketCellSwap(
	context unsafe.Pointer,
	head C.int,
	expectedPresent C.uint32_t,
	expected *C.uint8_t,
	expectedLength C.uint64_t,
	replacement *C.uint8_t,
	replacementLength C.uint64_t,
	outSwapped *C.uint32_t,
	failure *C.PotluckDbRuntimeHostErrorV1,
) C.int32_t {
	resetHostError(failure)
	bridge := bucketBridgeFor(context)
	if bridge == nil {
		return C.int32_t(BucketErrorContractViolation)
	}
	return bridge.guard("cell-swap", failure, func() C.int32_t {
		if outSwapped == nil {
			return bridge.fail(failure, &BucketError{
				Code:    BucketErrorContractViolation,
				Message: "bucket cell-swap output is null",
			})
		}
		if expectedPresent > 1 {
			return bridge.fail(failure, &BucketError{
				Code:    BucketErrorContractViolation,
				Message: "bucket expected-present flag is not Boolean",
			})
		}
		if expectedPresent == 0 && expectedLength != 0 {
			return bridge.fail(failure, &BucketError{
				Code:    BucketErrorContractViolation,
				Message: "bucket absent expected cell contains bytes",
			})
		}
		// A nil expected compares against absence; an empty non-nil one
		// compares against the empty value, and the two are different
		// questions under B2.
		var previous []byte
		if expectedPresent != 0 {
			previous = bucketBytes(expected, expectedLength)
		}
		var swapped bool
		var err error
		if head != 0 {
			swapped, err = bridge.storage.HeadSwap(
				previous, bucketBytes(replacement, replacementLength),
			)
		} else {
			swapped, err = bridge.storage.FloorSwap(
				previous, bucketBytes(replacement, replacementLength),
			)
		}
		if err != nil {
			return bridge.fail(failure, err)
		}
		if swapped {
			*outSwapped = 1
		} else {
			*outSwapped = 0
		}
		return 0
	})
}

//export potluckdbGoBucketCellPairGet
func potluckdbGoBucketCellPairGet(
	context unsafe.Pointer,
	emitHead C.PotluckDbRuntimeEmitBytesV1,
	headContext unsafe.Pointer,
	emitFloor C.PotluckDbRuntimeEmitBytesV1,
	floorContext unsafe.Pointer,
	failure *C.PotluckDbRuntimeHostErrorV1,
) C.int32_t {
	resetHostError(failure)
	bridge := bucketBridgeFor(context)
	if bridge == nil {
		return C.int32_t(BucketErrorContractViolation)
	}
	return bridge.guard("cell-pair-get", failure, func() C.int32_t {
		if bridge.cellPair == nil {
			return bridge.fail(failure, &BucketError{
				Code:    BucketErrorContractViolation,
				Message: "bucket declared no cell-pair capability",
			})
		}
		head, floor, err := bridge.cellPair.CellPairGet()
		if err != nil {
			return bridge.fail(failure, err)
		}
		if head != nil {
			if code := bucketEmit(emitHead, headContext, head); code != 0 {
				return code
			}
		}
		if floor != nil {
			if code := bucketEmit(emitFloor, floorContext, floor); code != 0 {
				return code
			}
		}
		return 0
	})
}

//export potluckdbGoBucketCounters
func potluckdbGoBucketCounters(
	context unsafe.Pointer,
	out *C.PotluckDbRuntimeBucketCounterSnapshotV1,
) {
	if out == nil {
		return
	}
	bridge := bucketBridgeFor(context)
	if bridge == nil {
		return
	}
	// A panic here has no error output to report through, so the snapshot
	// stays at the zeroed value the Runtime already holds.
	defer func() { _ = recover() }()
	value := bridge.counters.snapshot()
	out.struct_size = C.uint32_t(unsafe.Sizeof(*out))
	out.abi_version = C.POTLUCKDB_RUNTIME_ABI_VERSION_V1
	out.gets = C.uint64_t(value.Gets)
	out.puts = C.uint64_t(value.Puts)
	out.idempotent_puts = C.uint64_t(value.IdempotentPuts)
	out.exists = C.uint64_t(value.Exists)
	out.lists = C.uint64_t(value.Lists)
	out.deletes = C.uint64_t(value.Deletes)
	out.cell_gets = C.uint64_t(value.CellGets)
	out.cell_swaps = C.uint64_t(value.CellSwaps)
	out.cell_swap_wins = C.uint64_t(value.CellSwapWins)
	out.transport_gets = C.uint64_t(value.TransportGets)
	out.transport_puts = C.uint64_t(value.TransportPuts)
	out.transport_heads = C.uint64_t(value.TransportHeads)
	out.transport_lists = C.uint64_t(value.TransportLists)
	out.transport_deletes = C.uint64_t(value.TransportDeletes)
	out.transport_bytes_received = C.uint64_t(value.TransportBytesReceived)
	out.transport_bytes_submitted = C.uint64_t(value.TransportBytesSubmitted)
}

// StorageAdapterFor builds the Runtime bucket adapter table over a Go
// implementor, so RemoteFromStorage can take one directly.
//
// Ownership: the returned StorageAdapter owns the C table and the handle
// naming storage. The Runtime copies the table on open and keeps the
// context, so callbacks may arrive on any Runtime worker thread until the
// Remote that borrowed the adapter is closed and released. The adapter's
// C storage is reclaimed only once it is unreachable, which the Remote
// holding it prevents; nothing here frees it while a live Remote could
// still call back.
//
// Integrator API: for code that supplies its own storage transport. Application
// code reaches a bucket through FromS3, FromGCS and FromAzure.
func StorageAdapterFor(storage BucketStorage) (*StorageAdapter, error) {
	if storage == nil {
		return nil, invalidArgument(
			"bucket storage must not be nil",
			"Supply a bucket implementation and retry.",
		)
	}
	counters, err := newBucketCounters(storage)
	if err != nil {
		return nil, err
	}
	bridge := &bucketBridge{
		storage:  storage,
		counters: counters,
		messages: make(map[string]unsafe.Pointer),
	}
	if pair, ok := storage.(CellPairBucketStorage); ok {
		bridge.cellPair = pair
	}
	bridge.self = cgo.NewHandle(bridge)
	cellPair := C.int(0)
	if bridge.cellPair != nil {
		cellPair = C.int(1)
	}
	block := C.potluckdb_go_bucket_bridge_new(
		C.uintptr_t(bridge.self), cellPair,
	)
	if block == nil {
		bridge.self.Delete()
		return nil, internalFailure("potluckdb could not allocate a bucket adapter")
	}
	bridge.block = block
	adapter, err := RetainStorageAdapter(unsafe.Pointer(&block.adapter), bridge)
	if err != nil {
		bridge.release()
		return nil, err
	}
	// The finalizer sits on the StorageAdapter rather than on the bridge:
	// the cgo handle table holds the bridge, so the bridge is reachable
	// until the handle is deleted and could never be collected. The
	// Remote retains the adapter, and a finalized owner keeps what it
	// references alive for that collection, so the reclamation below runs
	// only after the Remote has released its native handle.
	runtime.SetFinalizer(adapter, func(*StorageAdapter) { bridge.release() })
	return adapter, nil
}

// RemoteFromBucketStorage opens a Remote over a Go bucket implementor;
// nil options select the defaults.
//
// Integrator API: for code that supplies its own storage transport. Application
// code reaches a bucket through FromS3, FromGCS and FromAzure.
func RemoteFromBucketStorage(
	storage BucketStorage, options *RemoteOptions,
) (*Remote, error) {
	adapter, err := StorageAdapterFor(storage)
	if err != nil {
		return nil, err
	}
	return RemoteFromStorage(adapter, options)
}
