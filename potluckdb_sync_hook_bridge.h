/* The A23 sync hook trampoline for Go.
 *
 * `OpenOptions.OnSync` reaches the Application ABI as a C function pointer
 * with a context, and Go cannot hand over the address of an //export'ed
 * function. This static trampoline has the ABI's hook signature and
 * forwards to the exported Go entry point, which is the only shape cgo
 * offers.
 *
 * Declarations only. This header is included by the preamble of a file
 * that uses //export, where cgo forbids definitions.
 */

#ifndef POTLUCKDB_GO_SYNC_HOOK_BRIDGE_H
#define POTLUCKDB_GO_SYNC_HOOK_BRIDGE_H

#include "potluckdb.h"

/* The cgo handle naming the Go hook, in C storage.
 *
 * The runtime keeps `on_sync_context` for the life of the session and
 * invokes the hook from the automatic-sync driver's thread, so the
 * context has to be C storage that outlives the open call. A Go pointer
 * here would break the cgo pointer rules the moment the runtime kept it.
 */
typedef struct PotluckDbGoSyncHookV1 {
    uintptr_t handle;
} PotluckDbGoSyncHookV1;

/* Allocate the context for one hook. Returns NULL when the allocation
 * fails. */
PotluckDbGoSyncHookV1 *potluckdb_go_sync_hook_new(uintptr_t handle);

/* Free a context. The caller has already established that no invocation
 * can still arrive: the database has closed. */
void potluckdb_go_sync_hook_free(PotluckDbGoSyncHookV1 *hook);

/* The function pointer the open options carry. */
PotluckDbOnSyncV1 potluckdb_go_sync_hook_function(void);

#endif /* POTLUCKDB_GO_SYNC_HOOK_BRIDGE_H */
