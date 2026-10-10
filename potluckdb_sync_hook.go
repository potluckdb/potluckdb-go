package potluckdb

/*
#include <stdlib.h>

#include "potluckdb_sync_hook_bridge.h"
*/
import "C"

import (
	"runtime/cgo"
	"unsafe"
)

// syncHook is the C storage and cgo handle behind one OpenOptions.OnSync.
//
// The runtime keeps the context for the life of the session and invokes
// the hook from the automatic-sync driver's thread, so both live until the
// database has closed. Close returns after the hook's last invocation, and
// the owner's afterClose releases them.
type syncHook struct {
	block *C.PotluckDbGoSyncHookV1
	self  cgo.Handle
}

func newSyncHook(hook func(SyncResult)) (*syncHook, error) {
	self := cgo.NewHandle(hook)
	block := C.potluckdb_go_sync_hook_new(C.uintptr_t(self))
	if block == nil {
		self.Delete()
		return nil, internalFailure("potluckdb could not allocate a sync hook")
	}
	return &syncHook{block: block, self: self}, nil
}

func (h *syncHook) fill(raw *C.PotluckDbOpenOptionsV1) {
	if h == nil {
		return
	}
	raw.on_sync = C.potluckdb_go_sync_hook_function()
	raw.on_sync_context = unsafe.Pointer(h.block)
}

func (h *syncHook) release() {
	if h == nil {
		return
	}
	C.potluckdb_go_sync_hook_free(h.block)
	h.block = nil
	h.self.Delete()
}

func syncHookFor(context unsafe.Pointer) (hook func(SyncResult)) {
	// A released handle makes Value panic, and a panic here would cross C
	// with no recover site above it. The database has closed by then, and
	// no invocation follows the close, so a missing hook is dropped.
	defer func() {
		if recover() != nil {
			hook = nil
		}
	}()
	if context == nil {
		return nil
	}
	handle := cgo.Handle(*(*C.uintptr_t)(context))
	hook, _ = handle.Value().(func(SyncResult))
	return hook
}

//export potluckdbGoOnSync
func potluckdbGoOnSync(context unsafe.Pointer, outcome *C.PotluckDbSyncResultV1) {
	result, err := decodeSyncResult(outcome)
	C.potluckdb_sync_result_release_v1(outcome)
	if err != nil {
		// An outcome this facade cannot read is still an attempt the hook
		// must hear about; it hears the facade's own failure.
		result = Failed{Err: err}
	}
	if hook := syncHookFor(context); hook != nil {
		hook(result)
	}
}
