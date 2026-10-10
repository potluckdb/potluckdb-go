/* Storage-library ABI as this SDK restates it.
 *
 * Loaded, never linked. The table create returns is the same
 * PotluckDbRuntimeBucketAdapterV1 RemoteFromStorage already takes; nothing
 * here is a Go callback. The struct layout is pinned by
 * tests/test_storage_pack.py against native/storage/src/abi.rs.
 */

#ifndef POTLUCKDB_GO_STORAGE_H
#define POTLUCKDB_GO_STORAGE_H

#include <stddef.h>
#include <stdint.h>

#include "potluckdb_runtime.h"

#define POTLUCKDB_STORAGE_ABI_VERSION_V1 UINT32_C(1)
#define POTLUCKDB_STORAGE_OK_V1 INT32_C(0)
#define POTLUCKDB_STORAGE_ERROR_V1 INT32_C(1)

/* This shim's own answer, not the library's: the named entry point is
 * not in the library that loaded. The library returns only OK or ERROR,
 * so a negative code cannot be confused with one of its refusals, and
 * the Go side words it -- a missing entry point is a stale install, not
 * a bad configuration. */
#define POTLUCKDB_GO_STORAGE_NO_ENTRY_POINT_V1 INT32_C(-1)

typedef struct PotluckDbStorageSliceV1 {
    const uint8_t *data;
    uint64_t length;
} PotluckDbStorageSliceV1;

typedef struct PotluckDbStorageErrorV1 {
    uint32_t struct_size;
    uint32_t abi_version;
    PotluckDbStorageSliceV1 message;
} PotluckDbStorageErrorV1;

/* Open the library once, from `path`, or from the platform's ordinary
 * name when `path` is empty, which lets the loader search. Where the
 * library is for this process is decided in potluckdb_storage.go. Never
 * unloaded: every table it handed out points into its code. */
int potluckdb_go_storage_load(const char *path, char *error, size_t error_length);

/* Build one provider's adapter, through the entry point `symbol` names.
 *
 * The symbol comes from the closed provider map in potluckdb_storage.go;
 * this shim resolves it and calls it, and answers
 * POTLUCKDB_GO_STORAGE_NO_ENTRY_POINT_V1 when the loaded library has no
 * such export. */
int32_t potluckdb_go_storage_create(
    const char *symbol,
    const uint8_t *json,
    uint64_t length,
    const PotluckDbRuntimeBucketAdapterV1 **out_adapter,
    char *error,
    size_t error_length
);

int32_t potluckdb_go_storage_release(
    const PotluckDbRuntimeBucketAdapterV1 *adapter
);

#endif /* POTLUCKDB_GO_STORAGE_H */
