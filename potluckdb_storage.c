#include "potluckdb_storage.h"

#include <stdlib.h>
#include <string.h>

#ifdef _WIN32
#include <windows.h>
#else
#include <dlfcn.h>
#endif

typedef uint32_t (*storage_abi_version_fn)(void);
typedef uint32_t (*storage_error_size_fn)(void);
typedef int32_t (*storage_error_release_fn)(PotluckDbStorageErrorV1 *);
typedef int32_t (*storage_create_fn)(
    PotluckDbStorageSliceV1,
    const PotluckDbRuntimeBucketAdapterV1 **,
    PotluckDbStorageErrorV1 *
);
typedef int32_t (*storage_release_fn)(const PotluckDbRuntimeBucketAdapterV1 *);

static void *storage_library;
static storage_release_fn storage_release;
static storage_error_release_fn storage_error_release;
static int storage_loaded;

static void copy_error(char *error, size_t error_length, const char *message) {
    size_t n;
    if (error == NULL || error_length == 0) {
        return;
    }
    if (message == NULL) {
        error[0] = '\0';
        return;
    }
    n = strlen(message);
    if (n >= error_length) {
        n = error_length - 1;
    }
    memcpy(error, message, n);
    error[n] = '\0';
}

static void *storage_symbol(void *handle, const char *name) {
#ifdef _WIN32
    return (void *)GetProcAddress((HMODULE)handle, name);
#else
    return dlsym(handle, name);
#endif
}

#ifdef _WIN32
static void *storage_load_windows(const char *path) {
    int needed;
    WCHAR *wide;
    void *handle;
    /* The ANSI loader interprets the path in the process code page,
     * which mangles non-ASCII user names; the Unicode loader takes
     * the UTF-16 this UTF-8 path converts to. The caller passes Go
     * strings, which are UTF-8. */
    needed = MultiByteToWideChar(
        CP_UTF8, MB_ERR_INVALID_CHARS, path, -1, NULL, 0
    );
    if (needed == 0) {
        return NULL;
    }
    wide = (WCHAR *)malloc((size_t)needed * sizeof(WCHAR));
    if (wide == NULL) {
        return NULL;
    }
    if (MultiByteToWideChar(
        CP_UTF8, MB_ERR_INVALID_CHARS, path, -1, wide, needed
    ) == 0) {
        free(wide);
        return NULL;
    }
    handle = (void *)LoadLibraryW(wide);
    free(wide);
    return handle;
}
#endif

int potluckdb_go_storage_load(const char *path, char *error, size_t error_length) {
    void *handle;
    storage_abi_version_fn abi_version;
    storage_error_size_fn error_size;
    storage_error_release_fn error_release;
    storage_release_fn release;
    uint32_t version;

    if (storage_loaded) {
        return 0;
    }
    if (path == NULL || path[0] == '\0') {
#ifdef _WIN32
        path = "potluckdb_storage.dll";
#elif defined(__APPLE__)
        path = "libpotluckdb_storage.dylib";
#else
        path = "libpotluckdb_storage.so";
#endif
    }

#ifdef _WIN32
    handle = storage_load_windows(path);
    if (handle == NULL) {
        copy_error(error, error_length, "the potluckdb storage library did not load");
        return -1;
    }
#else
    /* Cleared first: dlerror reports the last failure whenever it was. */
    (void)dlerror();
    handle = dlopen(path, RTLD_NOW);
    if (handle == NULL) {
        const char *reason = dlerror();
        if (reason == NULL) {
            reason = "no reason was reported";
        }
        copy_error(error, error_length, reason);
        return -1;
    }
#endif

    /* Every symbol every provider needs, before any of them is used. A
     * library that resolves three of four is not a usable library. The
     * create functions are not here: those are one per provider, named
     * by the closed map in potluckdb_storage.go and resolved below when
     * one is asked for. */
    abi_version = (storage_abi_version_fn)storage_symbol(
        handle, "potluckdb_storage_abi_version_v1"
    );
    error_size = (storage_error_size_fn)storage_symbol(
        handle, "potluckdb_storage_error_size_v1"
    );
    error_release = (storage_error_release_fn)storage_symbol(
        handle, "potluckdb_storage_error_release_v1"
    );
    release = (storage_release_fn)storage_symbol(
        handle, "potluckdb_storage_release_v1"
    );
    if (
        abi_version == NULL || error_size == NULL || error_release == NULL
        || release == NULL
    ) {
        copy_error(
            error,
            error_length,
            "the storage library does not export the v1 entry points"
        );
        return -1;
    }

    /* Checked before anything crosses, not after something goes wrong. */
    version = abi_version();
    if (version != POTLUCKDB_STORAGE_ABI_VERSION_V1) {
        copy_error(
            error, error_length, "the storage library speaks a different ABI"
        );
        return -1;
    }
    if (error_size() != (uint32_t)sizeof(PotluckDbStorageErrorV1)) {
        copy_error(
            error,
            error_length,
            "the storage library declares a different error structure"
        );
        return -1;
    }

    storage_library = handle;
    storage_release = release;
    storage_error_release = error_release;
    storage_loaded = 1;
    return 0;
}

int32_t potluckdb_go_storage_create(
    const char *symbol,
    const uint8_t *json,
    uint64_t length,
    const PotluckDbRuntimeBucketAdapterV1 **out_adapter,
    char *error,
    size_t error_length
) {
    PotluckDbStorageSliceV1 slice;
    PotluckDbStorageErrorV1 reported;
    storage_create_fn create;
    int32_t code;

    /* Loaded by potluckdb_go_storage_load first, which is where the Go
     * side decides which library this process uses; loading here by the
     * platform's name would go around that decision. */
    if (!storage_loaded) {
        copy_error(error, error_length, "the storage library has not been loaded");
        return POTLUCKDB_STORAGE_ERROR_V1;
    }
    if (out_adapter == NULL) {
        copy_error(error, error_length, "storage adapter out-parameter is null");
        return POTLUCKDB_STORAGE_ERROR_V1;
    }
    /* Resolved per provider rather than at load time: a library that
     * predates one provider is a working library for the others, and
     * refusing it wholesale would take S3 down with it. Absence is its
     * own answer -- Go words the refusal, because Go holds the map that
     * says which provider this symbol belongs to. */
    create = (storage_create_fn)storage_symbol(storage_library, symbol);
    if (create == NULL) {
        copy_error(error, error_length, "");
        return POTLUCKDB_GO_STORAGE_NO_ENTRY_POINT_V1;
    }
    *out_adapter = NULL;
    slice.data = json;
    slice.length = length;
    memset(&reported, 0, sizeof(reported));
    reported.struct_size = (uint32_t)sizeof(reported);
    reported.abi_version = POTLUCKDB_STORAGE_ABI_VERSION_V1;
    code = create(slice, out_adapter, &reported);
    if (code != POTLUCKDB_STORAGE_OK_V1) {
        if (reported.message.data != NULL && reported.message.length > 0) {
            size_t n = (size_t)reported.message.length;
            if (error != NULL && error_length > 0) {
                if (n >= error_length) {
                    n = error_length - 1;
                }
                memcpy(error, reported.message.data, n);
                error[n] = '\0';
            }
        } else {
            copy_error(
                error,
                error_length,
                "the storage library refused the configuration"
            );
        }
        (void)storage_error_release(&reported);
    }
    return code;
}

int32_t potluckdb_go_storage_release(
    const PotluckDbRuntimeBucketAdapterV1 *adapter
) {
    if (!storage_loaded || adapter == NULL) {
        return 1;
    }
    return storage_release(adapter);
}
