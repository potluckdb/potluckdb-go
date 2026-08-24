#ifndef POTLUCKDB_H
#define POTLUCKDB_H

#include "potluckdb_runtime.h"

#ifdef __cplusplus
extern "C" {
#endif

/*
 * Native Application ABI V1. Use this potluckdb_* surface for application
 * work; potluckdb_runtime_* is the lower integrator escape hatch.
 *
 * Input slices are borrowed for one call. Initialize every structure before
 * use. Release PotluckDbErrorV1 before reuse, Rows after borrowed values are no
 * longer read, and sync/conflict results with their matching release calls.
 * Close owner handles before release. A Database/Connection cannot close with
 * an active transaction, and a Remote cannot close while borrowed by an open
 * Database or Snapshot. Releasing an active transaction rolls it back.
 *
 * A default remote-backed Database owns its automatic-sync timer and worker;
 * callers and language SDKs must not start a second driver for that session.
 * Bucket adapter callbacks/context remain caller-owned and live until Remote
 * close. Concurrent calls follow contracts/application-abi.md.
 */

#define POTLUCKDB_C_SDK_VERSION_V1 UINT32_C(1)
#define POTLUCKDB_APPLICATION_ABI_VERSION_V1 UINT32_C(1)

#define POTLUCKDB_APPLICATION_OK_V1 INT32_C(0)
#define POTLUCKDB_APPLICATION_ERROR_V1 INT32_C(1)

#define POTLUCKDB_ERROR_CALLER_V1 UINT32_C(1)
#define POTLUCKDB_ERROR_TEMPORARY_V1 UINT32_C(2)
#define POTLUCKDB_ERROR_PERMANENT_V1 UINT32_C(3)
#define POTLUCKDB_ERROR_INTERNAL_V1 UINT32_C(4)

#define POTLUCKDB_SYNC_AUTOMATIC_V1 UINT32_C(1)
#define POTLUCKDB_SYNC_MANUAL_V1 UINT32_C(2)
#define POTLUCKDB_PREFETCH_AUTOMATIC_V1 UINT32_C(1)
#define POTLUCKDB_PREFETCH_DISABLED_V1 UINT32_C(2)
#define POTLUCKDB_CHECKPOINT_AUTOMATIC_V1 UINT32_C(1)
#define POTLUCKDB_CHECKPOINT_DISABLED_V1 UINT32_C(2)
#define POTLUCKDB_TRANSACTION_DEFERRED_V1 UINT32_C(1)
#define POTLUCKDB_TRANSACTION_IMMEDIATE_V1 UINT32_C(2)
#define POTLUCKDB_TRANSACTION_EXCLUSIVE_V1 UINT32_C(3)

#define POTLUCKDB_BINDINGS_POSITIONAL_V1 UINT32_C(1)
#define POTLUCKDB_BINDINGS_NAMED_V1 UINT32_C(2)
#define POTLUCKDB_VALUE_NULL_V1 UINT32_C(1)
#define POTLUCKDB_VALUE_BOOLEAN_V1 UINT32_C(2)
#define POTLUCKDB_VALUE_INTEGER_V1 UINT32_C(3)
#define POTLUCKDB_VALUE_REAL_V1 UINT32_C(4)
#define POTLUCKDB_VALUE_TEXT_V1 UINT32_C(5)
#define POTLUCKDB_VALUE_BLOB_V1 UINT32_C(6)

#define POTLUCKDB_SYNC_UP_TO_DATE_V1 UINT32_C(1)
#define POTLUCKDB_SYNC_PUSHED_V1 UINT32_C(2)
#define POTLUCKDB_SYNC_PULLED_V1 UINT32_C(3)
#define POTLUCKDB_SYNC_CONFLICT_V1 UINT32_C(4)

#define POTLUCKDB_SYNC_STATE_MANUAL_V1 UINT32_C(1)
#define POTLUCKDB_SYNC_STATE_UNBOUND_V1 UINT32_C(2)
#define POTLUCKDB_SYNC_STATE_SCHEDULED_V1 UINT32_C(3)
#define POTLUCKDB_SYNC_STATE_SYNCHRONIZING_V1 UINT32_C(4)
#define POTLUCKDB_SYNC_STATE_IDLE_V1 UINT32_C(5)
#define POTLUCKDB_SYNC_STATE_BACKOFF_V1 UINT32_C(6)
#define POTLUCKDB_SYNC_STATE_CONFLICT_V1 UINT32_C(7)
#define POTLUCKDB_SYNC_STATE_FAILED_V1 UINT32_C(8)
#define POTLUCKDB_SYNC_STATE_CLOSED_V1 UINT32_C(9)

#define POTLUCKDB_OPEN_ENCRYPTION_KEY_V1 UINT32_C(1)
#define POTLUCKDB_OPEN_CLIENT_ID_V1 UINT32_C(2)
#define POTLUCKDB_REMOTE_RETENTION_V1 UINT32_C(1)
#define POTLUCKDB_REENCRYPT_SOURCE_KEY_V1 UINT32_C(1)
#define POTLUCKDB_REENCRYPT_REVISION_V1 UINT32_C(2)

typedef PotluckDbRuntimeSliceV1 PotluckDbSliceV1;
typedef PotluckDbRuntimeBufferV1 PotluckDbBufferV1;
typedef uint64_t PotluckDbDatabaseHandleV1;
typedef uint64_t PotluckDbConnectionHandleV1;
typedef uint64_t PotluckDbSnapshotHandleV1;
typedef uint64_t PotluckDbRemoteHandleV1;
typedef uint64_t PotluckDbTransactionHandleV1;
typedef uint64_t PotluckDbRowsHandleV1;
typedef uint64_t PotluckDbStringListHandleV1;

typedef struct PotluckDbRevisionV1 {
    uint8_t bytes[32];
} PotluckDbRevisionV1;

typedef struct PotluckDbErrorV1 {
    uint32_t struct_size;
    uint32_t abi_version;
    uint32_t category;
    uint32_t has_retry_after;
    uint64_t retry_after_ms;
    PotluckDbBufferV1 code;
    PotluckDbBufferV1 message;
    PotluckDbBufferV1 action;
} PotluckDbErrorV1;

typedef struct PotluckDbSyncPolicyV1 {
    uint32_t struct_size;
    uint32_t mode;
    uint64_t outbound_debounce_ms;
    uint64_t active_pull_interval_ms;
    uint64_t idle_pull_interval_ms;
    uint64_t retry_initial_ms;
    uint64_t retry_max_ms;
} PotluckDbSyncPolicyV1;

typedef struct PotluckDbConnectionOptionsV1 {
    uint32_t struct_size;
    uint32_t prefetch_mode;
    uint32_t checkpoint_mode;
    uint32_t reserved;
    uint64_t timeout_ms;
    uint64_t maximum_rows;
    uint64_t prefetch_background_target_bytes;
    uint64_t prefetch_memory_budget_bytes;
    uint64_t checkpoint_soft_limit_bytes;
    uint64_t checkpoint_forced_limit_bytes;
} PotluckDbConnectionOptionsV1;

typedef struct PotluckDbOpenOptionsV1 {
    uint32_t struct_size;
    uint32_t flags;
    uint64_t local_budget_bytes;
    uint64_t minimum_local_budget_bytes;
    uint64_t local_compaction_bytes;
    uint8_t encryption_key[32];
    PotluckDbSliceV1 client_id;
    PotluckDbSyncPolicyV1 sync_policy;
    PotluckDbConnectionOptionsV1 connection;
} PotluckDbOpenOptionsV1;

typedef struct PotluckDbRetentionPolicyV1 {
    uint32_t struct_size;
    uint32_t reserved;
    uint64_t keep_last;
    uint64_t collect_at;
} PotluckDbRetentionPolicyV1;

typedef struct PotluckDbRemoteOptionsV1 {
    uint32_t struct_size;
    uint32_t flags;
    uint32_t read_concurrency;
    uint32_t write_concurrency;
    PotluckDbRetentionPolicyV1 retention;
} PotluckDbRemoteOptionsV1;

typedef struct PotluckDbValueV1 {
    uint32_t struct_size;
    uint32_t kind;
    uint32_t boolean_value;
    uint32_t reserved;
    int64_t integer_value;
    double real_value;
    PotluckDbSliceV1 bytes;
} PotluckDbValueV1;

typedef struct PotluckDbBindingV1 {
    uint32_t struct_size;
    uint32_t reserved;
    PotluckDbSliceV1 name;
    PotluckDbValueV1 value;
} PotluckDbBindingV1;

typedef struct PotluckDbBindingsV1 {
    uint32_t struct_size;
    uint32_t kind;
    const PotluckDbBindingV1 *items;
    uint64_t count;
} PotluckDbBindingsV1;

typedef struct PotluckDbExecuteResultV1 {
    uint32_t struct_size;
    uint32_t has_last_insert_id;
    int64_t rows_affected;
    int64_t last_insert_id;
} PotluckDbExecuteResultV1;

typedef struct PotluckDbDatabaseStatusV1 {
    uint32_t struct_size;
    uint32_t has_revision;
    PotluckDbRevisionV1 revision;
    uint64_t pending_writes;
    uint64_t local_bytes;
    uint64_t local_budget_bytes;
    uint64_t connection_count;
    uint32_t sync_state;
    uint32_t has_next_sync;
    uint64_t next_sync_in_ms;
    uint32_t has_conflict_revision;
    uint32_t closed;
    PotluckDbRevisionV1 conflict_revision;
} PotluckDbDatabaseStatusV1;

typedef struct PotluckDbConnectionStatusV1 {
    uint32_t struct_size;
    uint32_t transaction_active;
    uint32_t has_checkpoint;
    uint32_t has_prefetch;
    uint32_t closed;
    uint32_t reserved;
    PotluckDbRuntimeCheckpointStatusV1 checkpoint;
    PotluckDbRuntimePrefetchStatusV1 prefetch;
} PotluckDbConnectionStatusV1;

typedef struct PotluckDbSnapshotStatusV1 {
    uint32_t struct_size;
    uint32_t closed;
    PotluckDbRevisionV1 revision;
    uint64_t local_bytes;
} PotluckDbSnapshotStatusV1;

typedef struct PotluckDbRemoteStatusV1 {
    uint32_t struct_size;
    uint32_t has_retained_revisions;
    uint32_t maintenance_pending;
    uint32_t closed;
    uint64_t retained_revisions;
    uint64_t policy_failures;
} PotluckDbRemoteStatusV1;

typedef struct PotluckDbSyncResultV1 {
    uint32_t struct_size;
    uint32_t kind;
    uint32_t has_revision;
    uint32_t reserved;
    PotluckDbRevisionV1 revision;
    PotluckDbStringListHandleV1 replayed;
    PotluckDbStringListHandleV1 quarantined;
    PotluckDbBufferV1 reason;
} PotluckDbSyncResultV1;

typedef struct PotluckDbConflictResolutionV1 {
    uint32_t struct_size;
    uint32_t reserved;
    PotluckDbRevisionV1 revision;
    PotluckDbStringListHandleV1 already_published;
    PotluckDbStringListHandleV1 replayed;
    PotluckDbStringListHandleV1 quarantined;
    PotluckDbStringListHandleV1 pending;
} PotluckDbConflictResolutionV1;

typedef struct PotluckDbCollectionReportV1 {
    uint32_t struct_size;
    uint32_t has_floor_before;
    uint32_t has_floor_after;
    uint32_t aborted;
    PotluckDbRevisionV1 floor_before;
    PotluckDbRevisionV1 floor_after;
    uint64_t epoch;
    uint64_t marked;
    uint64_t swept;
} PotluckDbCollectionReportV1;

typedef struct PotluckDbCheckpointResultV1 {
    uint32_t struct_size;
    int32_t log_frames;
    int32_t checkpointed_frames;
    uint32_t reserved;
} PotluckDbCheckpointResultV1;

typedef struct PotluckDbReencryptOptionsV1 {
    uint32_t struct_size;
    uint32_t flags;
    uint8_t new_key[32];
    uint8_t source_key[32];
    PotluckDbRevisionV1 revision;
} PotluckDbReencryptOptionsV1;

typedef struct PotluckDbKeyRotationReportV1 {
    uint32_t struct_size;
    uint32_t has_source_key_id;
    uint32_t published;
    uint32_t reserved;
    PotluckDbRevisionV1 source_revision;
    PotluckDbRevisionV1 target_revision;
    PotluckDbRevisionV1 source_key_id;
    PotluckDbRevisionV1 target_key_id;
    uint64_t piece_count;
} PotluckDbKeyRotationReportV1;

uint32_t potluckdb_application_abi_version_v1(void);
uint32_t potluckdb_error_size_v1(void);
uint32_t potluckdb_sync_policy_size_v1(void);
uint32_t potluckdb_connection_options_size_v1(void);
uint32_t potluckdb_open_options_size_v1(void);
uint32_t potluckdb_retention_policy_size_v1(void);
uint32_t potluckdb_remote_options_size_v1(void);
uint32_t potluckdb_value_size_v1(void);
uint32_t potluckdb_binding_size_v1(void);
uint32_t potluckdb_bindings_size_v1(void);
uint32_t potluckdb_execute_result_size_v1(void);
uint32_t potluckdb_database_status_size_v1(void);
uint32_t potluckdb_connection_status_size_v1(void);
uint32_t potluckdb_snapshot_status_size_v1(void);
uint32_t potluckdb_remote_status_size_v1(void);
uint32_t potluckdb_sync_result_size_v1(void);
uint32_t potluckdb_conflict_resolution_size_v1(void);
uint32_t potluckdb_collection_report_size_v1(void);
uint32_t potluckdb_checkpoint_result_size_v1(void);
uint32_t potluckdb_reencrypt_options_size_v1(void);
uint32_t potluckdb_key_rotation_report_size_v1(void);

void potluckdb_error_init_v1(PotluckDbErrorV1 *value);
void potluckdb_sync_policy_init_v1(PotluckDbSyncPolicyV1 *value);
void potluckdb_connection_options_init_v1(PotluckDbConnectionOptionsV1 *value);
void potluckdb_open_options_init_v1(PotluckDbOpenOptionsV1 *value);
void potluckdb_retention_policy_init_v1(PotluckDbRetentionPolicyV1 *value);
void potluckdb_remote_options_init_v1(PotluckDbRemoteOptionsV1 *value);
void potluckdb_value_init_v1(PotluckDbValueV1 *value);
void potluckdb_binding_init_v1(PotluckDbBindingV1 *value);
void potluckdb_bindings_init_v1(PotluckDbBindingsV1 *value);
void potluckdb_execute_result_init_v1(PotluckDbExecuteResultV1 *value);
void potluckdb_database_status_init_v1(PotluckDbDatabaseStatusV1 *value);
void potluckdb_connection_status_init_v1(PotluckDbConnectionStatusV1 *value);
void potluckdb_snapshot_status_init_v1(PotluckDbSnapshotStatusV1 *value);
void potluckdb_remote_status_init_v1(PotluckDbRemoteStatusV1 *value);
void potluckdb_sync_result_init_v1(PotluckDbSyncResultV1 *value);
void potluckdb_conflict_resolution_init_v1(PotluckDbConflictResolutionV1 *value);
void potluckdb_collection_report_init_v1(PotluckDbCollectionReportV1 *value);
void potluckdb_checkpoint_result_init_v1(PotluckDbCheckpointResultV1 *value);
void potluckdb_reencrypt_options_init_v1(PotluckDbReencryptOptionsV1 *value);
void potluckdb_key_rotation_report_init_v1(PotluckDbKeyRotationReportV1 *value);

int32_t potluckdb_error_release_v1(PotluckDbErrorV1 *value);
int32_t potluckdb_buffer_release_v1(PotluckDbBufferV1 *value);
int32_t potluckdb_revision_parse_v1(PotluckDbSliceV1 text, PotluckDbRevisionV1 *out_revision, PotluckDbErrorV1 *out_error);
/* Format owns out_text; release it with potluckdb_buffer_release_v1. */
int32_t potluckdb_revision_format_v1(const PotluckDbRevisionV1 *revision, PotluckDbBufferV1 *out_text, PotluckDbErrorV1 *out_error);

int32_t potluckdb_remote_from_directory_v1(PotluckDbSliceV1 path, const PotluckDbRemoteOptionsV1 *options, PotluckDbRemoteHandleV1 *out_remote, PotluckDbErrorV1 *out_error);
int32_t potluckdb_remote_from_storage_v1(const PotluckDbRuntimeBucketAdapterV1 *storage, const PotluckDbRemoteOptionsV1 *options, PotluckDbRemoteHandleV1 *out_remote, PotluckDbErrorV1 *out_error);
int32_t potluckdb_remote_status_v1(PotluckDbRemoteHandleV1 remote, PotluckDbRemoteStatusV1 *out_status, uint32_t out_status_size, PotluckDbErrorV1 *out_error);
int32_t potluckdb_remote_collect_v1(PotluckDbRemoteHandleV1 remote, const PotluckDbRetentionPolicyV1 *retention, PotluckDbCollectionReportV1 *out_report, uint32_t out_report_size, PotluckDbErrorV1 *out_error);
int32_t potluckdb_remote_recover_abandoned_collection_v1(PotluckDbRemoteHandleV1 remote, PotluckDbCollectionReportV1 *out_report, uint32_t out_report_size, PotluckDbErrorV1 *out_error);
int32_t potluckdb_remote_close_v1(PotluckDbRemoteHandleV1 remote, PotluckDbErrorV1 *out_error);
int32_t potluckdb_remote_release_v1(PotluckDbRemoteHandleV1 remote, PotluckDbErrorV1 *out_error);

int32_t potluckdb_open_v1(PotluckDbSliceV1 path, PotluckDbRemoteHandleV1 remote, const PotluckDbOpenOptionsV1 *options, PotluckDbDatabaseHandleV1 *out_database, PotluckDbErrorV1 *out_error);
int32_t potluckdb_open_revision_v1(PotluckDbSliceV1 path, const PotluckDbRevisionV1 *revision, PotluckDbRemoteHandleV1 remote, const PotluckDbOpenOptionsV1 *options, PotluckDbSnapshotHandleV1 *out_snapshot, PotluckDbErrorV1 *out_error);

int32_t potluckdb_database_query_v1(PotluckDbDatabaseHandleV1 database, PotluckDbSliceV1 sql, const PotluckDbBindingsV1 *bindings, PotluckDbRowsHandleV1 *out_rows, PotluckDbErrorV1 *out_error);
int32_t potluckdb_database_execute_v1(PotluckDbDatabaseHandleV1 database, PotluckDbSliceV1 sql, const PotluckDbBindingsV1 *bindings, PotluckDbExecuteResultV1 *out_result, uint32_t out_result_size, PotluckDbErrorV1 *out_error);
int32_t potluckdb_database_transaction_begin_v1(PotluckDbDatabaseHandleV1 database, uint32_t mode, PotluckDbTransactionHandleV1 *out_transaction, PotluckDbErrorV1 *out_error);
int32_t potluckdb_database_connect_v1(PotluckDbDatabaseHandleV1 database, const PotluckDbConnectionOptionsV1 *options, PotluckDbConnectionHandleV1 *out_connection, PotluckDbErrorV1 *out_error);
int32_t potluckdb_database_sync_v1(PotluckDbDatabaseHandleV1 database, PotluckDbSyncResultV1 *out_result, uint32_t out_result_size, PotluckDbErrorV1 *out_error);
int32_t potluckdb_database_resolve_conflict_v1(PotluckDbDatabaseHandleV1 database, PotluckDbConflictResolutionV1 *out_result, uint32_t out_result_size, PotluckDbErrorV1 *out_error);
int32_t potluckdb_database_set_sync_policy_v1(PotluckDbDatabaseHandleV1 database, const PotluckDbSyncPolicyV1 *policy, PotluckDbDatabaseStatusV1 *out_status, uint32_t out_status_size, PotluckDbErrorV1 *out_error);
int32_t potluckdb_database_clear_quarantine_records_v1(PotluckDbDatabaseHandleV1 database, uint64_t *out_count, PotluckDbErrorV1 *out_error);
int32_t potluckdb_database_status_v1(PotluckDbDatabaseHandleV1 database, PotluckDbDatabaseStatusV1 *out_status, uint32_t out_status_size, PotluckDbErrorV1 *out_error);
int32_t potluckdb_database_close_v1(PotluckDbDatabaseHandleV1 database, PotluckDbErrorV1 *out_error);
int32_t potluckdb_database_release_v1(PotluckDbDatabaseHandleV1 database, PotluckDbErrorV1 *out_error);

int32_t potluckdb_connection_query_v1(PotluckDbConnectionHandleV1 connection, PotluckDbSliceV1 sql, const PotluckDbBindingsV1 *bindings, PotluckDbRowsHandleV1 *out_rows, PotluckDbErrorV1 *out_error);
int32_t potluckdb_connection_execute_v1(PotluckDbConnectionHandleV1 connection, PotluckDbSliceV1 sql, const PotluckDbBindingsV1 *bindings, PotluckDbExecuteResultV1 *out_result, uint32_t out_result_size, PotluckDbErrorV1 *out_error);
int32_t potluckdb_connection_transaction_begin_v1(PotluckDbConnectionHandleV1 connection, uint32_t mode, PotluckDbTransactionHandleV1 *out_transaction, PotluckDbErrorV1 *out_error);
int32_t potluckdb_connection_checkpoint_v1(PotluckDbConnectionHandleV1 connection, PotluckDbCheckpointResultV1 *out_result, uint32_t out_result_size, PotluckDbErrorV1 *out_error);
int32_t potluckdb_connection_status_v1(PotluckDbConnectionHandleV1 connection, PotluckDbConnectionStatusV1 *out_status, uint32_t out_status_size, PotluckDbErrorV1 *out_error);
int32_t potluckdb_connection_close_v1(PotluckDbConnectionHandleV1 connection, PotluckDbErrorV1 *out_error);
int32_t potluckdb_connection_release_v1(PotluckDbConnectionHandleV1 connection, PotluckDbErrorV1 *out_error);

int32_t potluckdb_snapshot_query_v1(PotluckDbSnapshotHandleV1 snapshot, PotluckDbSliceV1 sql, const PotluckDbBindingsV1 *bindings, PotluckDbRowsHandleV1 *out_rows, PotluckDbErrorV1 *out_error);
int32_t potluckdb_snapshot_status_v1(PotluckDbSnapshotHandleV1 snapshot, PotluckDbSnapshotStatusV1 *out_status, uint32_t out_status_size, PotluckDbErrorV1 *out_error);
int32_t potluckdb_snapshot_close_v1(PotluckDbSnapshotHandleV1 snapshot, PotluckDbErrorV1 *out_error);
int32_t potluckdb_snapshot_release_v1(PotluckDbSnapshotHandleV1 snapshot, PotluckDbErrorV1 *out_error);

int32_t potluckdb_transaction_query_v1(PotluckDbTransactionHandleV1 transaction, PotluckDbSliceV1 sql, const PotluckDbBindingsV1 *bindings, PotluckDbRowsHandleV1 *out_rows, PotluckDbErrorV1 *out_error);
int32_t potluckdb_transaction_execute_v1(PotluckDbTransactionHandleV1 transaction, PotluckDbSliceV1 sql, const PotluckDbBindingsV1 *bindings, PotluckDbExecuteResultV1 *out_result, uint32_t out_result_size, PotluckDbErrorV1 *out_error);
int32_t potluckdb_transaction_commit_v1(PotluckDbTransactionHandleV1 transaction, PotluckDbErrorV1 *out_error);
int32_t potluckdb_transaction_rollback_v1(PotluckDbTransactionHandleV1 transaction, PotluckDbErrorV1 *out_error);
int32_t potluckdb_transaction_release_v1(PotluckDbTransactionHandleV1 transaction, PotluckDbErrorV1 *out_error);

int32_t potluckdb_rows_row_count_v1(PotluckDbRowsHandleV1 rows, uint64_t *out_count, PotluckDbErrorV1 *out_error);
int32_t potluckdb_rows_column_count_v1(PotluckDbRowsHandleV1 rows, uint64_t row, uint64_t *out_count, PotluckDbErrorV1 *out_error);
int32_t potluckdb_rows_value_v1(PotluckDbRowsHandleV1 rows, uint64_t row, uint64_t column, PotluckDbValueV1 *out_value, uint32_t out_value_size, PotluckDbErrorV1 *out_error);
int32_t potluckdb_rows_release_v1(PotluckDbRowsHandleV1 rows, PotluckDbErrorV1 *out_error);
int32_t potluckdb_string_list_count_v1(PotluckDbStringListHandleV1 list, uint64_t *out_count, PotluckDbErrorV1 *out_error);
int32_t potluckdb_string_list_get_v1(PotluckDbStringListHandleV1 list, uint64_t index, PotluckDbSliceV1 *out_value, PotluckDbErrorV1 *out_error);
int32_t potluckdb_sync_result_release_v1(PotluckDbSyncResultV1 *result);
int32_t potluckdb_conflict_resolution_release_v1(PotluckDbConflictResolutionV1 *result);

int32_t potluckdb_admin_reencrypt_v1(PotluckDbRemoteHandleV1 source, PotluckDbRemoteHandleV1 target, const PotluckDbReencryptOptionsV1 *options, PotluckDbKeyRotationReportV1 *out_report, uint32_t out_report_size, PotluckDbErrorV1 *out_error);

#define POTLUCKDB_AUTO_SYNC_DRIVER_RUNNING_V1 UINT32_C(1)
#define POTLUCKDB_AUTO_SYNC_DRIVER_PAUSED_V1 UINT32_C(2)
#define POTLUCKDB_AUTO_SYNC_DRIVER_STOPPED_V1 UINT32_C(3)
#define POTLUCKDB_AUTO_SYNC_DRIVER_FAILED_V1 UINT32_C(4)

typedef uint64_t PotluckDbAutoSyncDriverHandleV1;

typedef struct PotluckDbAutoSyncDriverStatusV1 {
    uint32_t struct_size;
    uint32_t sdk_version;
    uint32_t state;
    int32_t last_error_code;
    PotluckDbRuntimeHandleV1 session_handle;
    uint64_t wake_requests;
    uint64_t drive_calls;
    PotluckDbRuntimeBufferV1 last_error;
} PotluckDbAutoSyncDriverStatusV1;

uint32_t potluckdb_c_sdk_version_v1(void);
uint32_t potluckdb_auto_sync_driver_status_size_v1(void);
void potluckdb_auto_sync_driver_status_init_v1(
    PotluckDbAutoSyncDriverStatusV1 *status
);

/* Start one host timer/worker for one open, remote-bound automatic session. */
int32_t potluckdb_auto_sync_driver_start_v1(
    PotluckDbRuntimeHandleV1 session_handle,
    PotluckDbAutoSyncDriverHandleV1 *out_driver,
    PotluckDbRuntimeErrorV1 *out_error
);
/* Wake after a call that may commit, change policy, or resolve a conflict. */
int32_t potluckdb_auto_sync_driver_wake_v1(
    PotluckDbAutoSyncDriverHandleV1 driver_handle,
    PotluckDbRuntimeErrorV1 *out_error
);
/* Pause waits for an active drive to quiesce; resume schedules a fresh drive. */
int32_t potluckdb_auto_sync_driver_pause_v1(
    PotluckDbAutoSyncDriverHandleV1 driver_handle,
    PotluckDbRuntimeErrorV1 *out_error
);
int32_t potluckdb_auto_sync_driver_resume_v1(
    PotluckDbAutoSyncDriverHandleV1 driver_handle,
    PotluckDbRuntimeErrorV1 *out_error
);
/* Status owns last_error; release its nonzero token with potluckdb_runtime_buffer_release_v1. */
int32_t potluckdb_auto_sync_driver_status_v1(
    PotluckDbAutoSyncDriverHandleV1 driver_handle,
    PotluckDbAutoSyncDriverStatusV1 *out_status,
    uint32_t out_status_size,
    PotluckDbRuntimeErrorV1 *out_error
);
/* Stop joins the worker. Stop, then release, before closing the session. */
int32_t potluckdb_auto_sync_driver_stop_v1(
    PotluckDbAutoSyncDriverHandleV1 driver_handle,
    PotluckDbRuntimeErrorV1 *out_error
);
int32_t potluckdb_auto_sync_driver_release_v1(
    PotluckDbAutoSyncDriverHandleV1 driver_handle,
    PotluckDbRuntimeErrorV1 *out_error
);

#ifdef __cplusplus
}
#endif

#endif
