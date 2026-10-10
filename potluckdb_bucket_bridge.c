#include <stdlib.h>

#include "potluckdb_bucket_bridge.h"
#include "_cgo_export.h"

extern uint32_t potluckdb_internal_storage_callback_active_v1(void);
int potluckdb_go_in_bucket_callback(void) {
    return potluckdb_internal_storage_callback_active_v1() != 0;
}

/* The cast away from `const` is cgo's: it declares every exported
 * pointer parameter mutable, and the seam hands these slices in const. */

static int32_t bucket_get(
    void *context,
    PotluckDbRuntimeSliceV1 key,
    PotluckDbRuntimeEmitBytesV1 emit,
    void *emit_context,
    PotluckDbRuntimeHostErrorV1 *error
) {
    return potluckdbGoBucketGet(
        context, (uint8_t *)key.data, key.length, emit, emit_context, error
    );
}

static int32_t bucket_put(
    void *context,
    PotluckDbRuntimeSliceV1 key,
    PotluckDbRuntimeSliceV1 data,
    PotluckDbRuntimeHostErrorV1 *error
) {
    return potluckdbGoBucketPut(
        context,
        (uint8_t *)key.data,
        key.length,
        (uint8_t *)data.data,
        data.length,
        error
    );
}

static int32_t bucket_exists(
    void *context,
    PotluckDbRuntimeSliceV1 key,
    uint32_t *out_exists,
    PotluckDbRuntimeHostErrorV1 *error
) {
    return potluckdbGoBucketExists(
        context, (uint8_t *)key.data, key.length, out_exists, error
    );
}

static int32_t bucket_list(
    void *context,
    PotluckDbRuntimeSliceV1 prefix,
    PotluckDbRuntimeEmitBytesV1 emit,
    void *emit_context,
    PotluckDbRuntimeHostErrorV1 *error
) {
    return potluckdbGoBucketList(
        context, (uint8_t *)prefix.data, prefix.length, emit, emit_context, error
    );
}

static int32_t bucket_delete(
    void *context,
    PotluckDbRuntimeSliceV1 key,
    PotluckDbRuntimeHostErrorV1 *error
) {
    return potluckdbGoBucketDelete(
        context, (uint8_t *)key.data, key.length, error
    );
}

static int32_t bucket_head_get(
    void *context,
    PotluckDbRuntimeEmitBytesV1 emit,
    void *emit_context,
    PotluckDbRuntimeHostErrorV1 *error
) {
    return potluckdbGoBucketCellGet(context, 1, emit, emit_context, error);
}

static int32_t bucket_floor_get(
    void *context,
    PotluckDbRuntimeEmitBytesV1 emit,
    void *emit_context,
    PotluckDbRuntimeHostErrorV1 *error
) {
    return potluckdbGoBucketCellGet(context, 0, emit, emit_context, error);
}

static int32_t bucket_head_swap(
    void *context,
    uint32_t expected_present,
    PotluckDbRuntimeSliceV1 expected,
    PotluckDbRuntimeSliceV1 new_value,
    uint32_t *out_swapped,
    PotluckDbRuntimeHostErrorV1 *error
) {
    return potluckdbGoBucketCellSwap(
        context,
        1,
        expected_present,
        (uint8_t *)expected.data,
        expected.length,
        (uint8_t *)new_value.data,
        new_value.length,
        out_swapped,
        error
    );
}

static int32_t bucket_floor_swap(
    void *context,
    uint32_t expected_present,
    PotluckDbRuntimeSliceV1 expected,
    PotluckDbRuntimeSliceV1 new_value,
    uint32_t *out_swapped,
    PotluckDbRuntimeHostErrorV1 *error
) {
    return potluckdbGoBucketCellSwap(
        context,
        0,
        expected_present,
        (uint8_t *)expected.data,
        expected.length,
        (uint8_t *)new_value.data,
        new_value.length,
        out_swapped,
        error
    );
}

static int32_t bucket_cell_pair_get(
    void *context,
    PotluckDbRuntimeEmitBytesV1 emit_head,
    void *head_context,
    PotluckDbRuntimeEmitBytesV1 emit_floor,
    void *floor_context,
    PotluckDbRuntimeHostErrorV1 *error
) {
    return potluckdbGoBucketCellPairGet(
        context, emit_head, head_context, emit_floor, floor_context, error
    );
}

static void bucket_counters(
    void *context,
    PotluckDbRuntimeBucketCounterSnapshotV1 *out_counters
) {
    potluckdbGoBucketCounters(context, out_counters);
}

PotluckDbGoBucketBridgeV1 *potluckdb_go_bucket_bridge_new(
    uintptr_t handle,
    int cell_pair
) {
    PotluckDbGoBucketBridgeV1 *bridge =
        (PotluckDbGoBucketBridgeV1 *)calloc(1, sizeof(PotluckDbGoBucketBridgeV1));
    if (bridge == NULL) {
        return NULL;
    }
    /* Filled in rather than initialized through the Runtime Interface: an
     * application SDK calls the Application ABI and nothing below it, and
     * calloc has already zeroed every field these two do not name. */
    bridge->adapter.struct_size = (uint32_t)sizeof(PotluckDbRuntimeBucketAdapterV1);
    bridge->adapter.abi_version = POTLUCKDB_RUNTIME_ABI_VERSION_V1;
    bridge->handle = handle;
    bridge->adapter.context = &bridge->handle;
    bridge->adapter.get = bucket_get;
    bridge->adapter.put = bucket_put;
    bridge->adapter.exists = bucket_exists;
    bridge->adapter.list = bucket_list;
    bridge->adapter.delete_value = bucket_delete;
    bridge->adapter.head_get = bucket_head_get;
    bridge->adapter.head_swap = bucket_head_swap;
    bridge->adapter.floor_get = bucket_floor_get;
    bridge->adapter.floor_swap = bucket_floor_swap;
    bridge->adapter.cell_pair_get = cell_pair ? bucket_cell_pair_get : NULL;
    bridge->adapter.counters = bucket_counters;
    return bridge;
}

void potluckdb_go_bucket_bridge_free(PotluckDbGoBucketBridgeV1 *bridge) {
    free(bridge);
}

int32_t potluckdb_go_bucket_emit(
    PotluckDbRuntimeEmitBytesV1 emit,
    void *emit_context,
    const uint8_t *data,
    uint64_t length
) {
    PotluckDbRuntimeSliceV1 value;
    if (emit == NULL) {
        return POTLUCKDB_RUNTIME_ERROR_CONTRACT_VIOLATION;
    }
    value.data = data;
    value.length = length;
    return emit(emit_context, value);
}
