#include <stdlib.h>

#include "potluckdb_sync_hook_bridge.h"
#include "_cgo_export.h"

/* The outcome arrives by value and belongs to the hook; Go decodes and
 * releases it through the pointer before this frame returns. */
static void sync_hook(void *context, PotluckDbSyncResultV1 outcome) {
    potluckdbGoOnSync(context, &outcome);
}

PotluckDbGoSyncHookV1 *potluckdb_go_sync_hook_new(uintptr_t handle) {
    PotluckDbGoSyncHookV1 *hook =
        (PotluckDbGoSyncHookV1 *)calloc(1, sizeof(PotluckDbGoSyncHookV1));
    if (hook == NULL) {
        return NULL;
    }
    hook->handle = handle;
    return hook;
}

void potluckdb_go_sync_hook_free(PotluckDbGoSyncHookV1 *hook) {
    free(hook);
}

PotluckDbOnSyncV1 potluckdb_go_sync_hook_function(void) {
    return sync_hook;
}
