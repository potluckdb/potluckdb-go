/* Seam-4 trampolines for the Go bucket bridge.
 *
 * The table the Runtime Interface takes is a set of C function pointers,
 * and Go cannot produce one: an //export'ed Go function is reachable from
 * C but its address is not a value Go code can hand over. These static
 * trampolines have the exact seam signatures and forward to the exported
 * Go entry points, which is the only shape cgo offers.
 *
 * Declarations only. This header is included by the preamble of a file
 * that uses //export, where cgo forbids definitions.
 */

#ifndef POTLUCKDB_GO_BUCKET_BRIDGE_H
#define POTLUCKDB_GO_BUCKET_BRIDGE_H

#include "potluckdb_runtime.h"

int potluckdb_go_in_bucket_callback(void);

/* One bucket adapter plus the cgo handle naming its Go implementor.
 *
 * The Runtime copies the table and keeps the context pointer, so the
 * context has to be C storage that outlives the call: `handle` is that
 * storage, and `adapter.context` points at it. A Go pointer here would
 * break the cgo pointer rules the moment the Runtime retained it.
 */
typedef struct PotluckDbGoBucketBridgeV1 {
    PotluckDbRuntimeBucketAdapterV1 adapter;
    uintptr_t handle;
} PotluckDbGoBucketBridgeV1;

/* Allocate a bridge whose table points at the trampolines below.
 * `cell_pair` selects the optional B13 capability. Returns NULL when the
 * allocation fails. */
PotluckDbGoBucketBridgeV1 *potluckdb_go_bucket_bridge_new(
    uintptr_t handle,
    int cell_pair
);

/* Free a bridge. The caller has already established that no callback can
 * still arrive. */
void potluckdb_go_bucket_bridge_free(PotluckDbGoBucketBridgeV1 *bridge);

/* Call the Runtime's emit callback with one value. Go cannot call a C
 * function pointer, so emitting goes through here. */
int32_t potluckdb_go_bucket_emit(
    PotluckDbRuntimeEmitBytesV1 emit,
    void *emit_context,
    const uint8_t *data,
    uint64_t length
);

#endif /* POTLUCKDB_GO_BUCKET_BRIDGE_H */
