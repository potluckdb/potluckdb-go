#ifndef POTLUCKDB_RUNTIME_H
#define POTLUCKDB_RUNTIME_H

#include <stdint.h>

#ifdef __cplusplus
extern "C" {
#endif

#define POTLUCKDB_RUNTIME_ABI_VERSION_V1 UINT32_C(1)

#define POTLUCKDB_RUNTIME_OK INT32_C(0)
#define POTLUCKDB_RUNTIME_ERROR_INTERNAL_FAILURE INT32_C(1)
#define POTLUCKDB_RUNTIME_ERROR_MISSING INT32_C(2)
#define POTLUCKDB_RUNTIME_ERROR_STORAGE_FULL INT32_C(3)
#define POTLUCKDB_RUNTIME_ERROR_STORAGE_UNAVAILABLE INT32_C(4)
#define POTLUCKDB_RUNTIME_ERROR_INTEGRITY INT32_C(5)
#define POTLUCKDB_RUNTIME_ERROR_CONFLICT INT32_C(6)
#define POTLUCKDB_RUNTIME_ERROR_UNAVAILABLE INT32_C(7)
#define POTLUCKDB_RUNTIME_ERROR_FORMAT_VERSION_UNSUPPORTED INT32_C(8)
#define POTLUCKDB_RUNTIME_ERROR_KEY_REQUIRED INT32_C(9)
#define POTLUCKDB_RUNTIME_ERROR_WRONG_KEY_OR_CORRUPT INT32_C(10)
#define POTLUCKDB_RUNTIME_ERROR_REMOTE_READ_REQUIRED INT32_C(11)
#define POTLUCKDB_RUNTIME_ERROR_READ_ONLY_REVISION INT32_C(12)
#define POTLUCKDB_RUNTIME_ERROR_HISTORY_RETIRED INT32_C(13)
#define POTLUCKDB_RUNTIME_ERROR_SYNC_DEFERRED INT32_C(14)
#define POTLUCKDB_RUNTIME_ERROR_WRITE_CAPACITY_EXCEEDED INT32_C(15)
#define POTLUCKDB_RUNTIME_ERROR_BUDGET_REFUSED INT32_C(16)
#define POTLUCKDB_RUNTIME_ERROR_IDENTITY_REFUSED INT32_C(17)
#define POTLUCKDB_RUNTIME_ERROR_CONTRACT_VIOLATION INT32_C(18)
#define POTLUCKDB_RUNTIME_ERROR_IDENTITY_COLLISION INT32_C(19)
#define POTLUCKDB_RUNTIME_ERROR_SQLITE INT32_C(20)
#define POTLUCKDB_RUNTIME_ERROR_EXECUTION_ENGINE INT32_C(21)

#define POTLUCKDB_RUNTIME_DUCKDB_VALUE_BOOLEAN_V1 UINT8_C(1)
#define POTLUCKDB_RUNTIME_DUCKDB_VALUE_SIGNED_INTEGER_V1 UINT8_C(2)
#define POTLUCKDB_RUNTIME_DUCKDB_VALUE_UNSIGNED_INTEGER_V1 UINT8_C(3)
#define POTLUCKDB_RUNTIME_DUCKDB_VALUE_REAL_V1 UINT8_C(4)
#define POTLUCKDB_RUNTIME_DUCKDB_VALUE_TEXT_V1 UINT8_C(5)
#define POTLUCKDB_RUNTIME_DUCKDB_VALUE_BLOB_V1 UINT8_C(6)

#define POTLUCKDB_RUNTIME_PROFILE_ROLLBACK_V1 UINT32_C(1)
#define POTLUCKDB_RUNTIME_PROFILE_WAL_EXCLUSIVE_V1 UINT32_C(2)

#define POTLUCKDB_RUNTIME_OPEN_ENCRYPTION_KEY UINT32_C(1)
#define POTLUCKDB_RUNTIME_SPARSE_EXACT_REVISION_V1 UINT32_C(1)
#define POTLUCKDB_RUNTIME_PREFETCH_FIXED_V1 UINT32_C(1)
#define POTLUCKDB_RUNTIME_PREFETCH_ADAPTIVE_V1 UINT32_C(2)
#define POTLUCKDB_RUNTIME_PREFETCH_STATEMENT_V1 UINT32_C(1)
#define POTLUCKDB_RUNTIME_PREFETCH_BACKGROUND_V1 UINT32_C(2)
#define POTLUCKDB_RUNTIME_PREFETCH_ADAPTIVE_STARTING_V1 UINT32_C(4)
#define POTLUCKDB_RUNTIME_PREFETCH_DISABLED_V1 UINT32_C(0)
#define POTLUCKDB_RUNTIME_PREFETCH_IDLE_V1 UINT32_C(1)
#define POTLUCKDB_RUNTIME_PREFETCH_FETCHING_V1 UINT32_C(2)
#define POTLUCKDB_RUNTIME_PREFETCH_INSTALLING_V1 UINT32_C(3)
#define POTLUCKDB_RUNTIME_PREFETCH_PAUSED_V1 UINT32_C(4)
#define POTLUCKDB_RUNTIME_PREFETCH_COMPLETE_V1 UINT32_C(5)
#define POTLUCKDB_RUNTIME_PREFETCH_SATURATED_V1 UINT32_C(6)
#define POTLUCKDB_RUNTIME_PREFETCH_FAILED_V1 UINT32_C(7)
#define POTLUCKDB_RUNTIME_PREFETCH_CLOSED_V1 UINT32_C(8)
#define POTLUCKDB_RUNTIME_ROTATION_SOURCE_KEY_V1 UINT32_C(1)
#define POTLUCKDB_RUNTIME_ROTATION_REVISION_V1 UINT32_C(2)

#define POTLUCKDB_RUNTIME_UPLOAD_FIXED_V1 UINT32_C(1)
#define POTLUCKDB_RUNTIME_UPLOAD_ADAPTIVE_V1 UINT32_C(2)
#define POTLUCKDB_RUNTIME_UPLOAD_CUSTOM_V1 UINT32_C(3)
#define POTLUCKDB_RUNTIME_UPLOAD_ADAPTIVE_STARTING_V1 UINT32_C(1)

#define POTLUCKDB_RUNTIME_TRANSFER_SUCCESS_V1 UINT32_C(1)
#define POTLUCKDB_RUNTIME_TRANSFER_UNAVAILABLE_V1 UINT32_C(2)
#define POTLUCKDB_RUNTIME_TRANSFER_ERROR_V1 UINT32_C(3)

#define POTLUCKDB_RUNTIME_WAL_VALIDATION_TRUST_SQLITE_V1 UINT32_C(1)
#define POTLUCKDB_RUNTIME_WAL_VALIDATION_STRICT_V1 UINT32_C(2)
#define POTLUCKDB_RUNTIME_ATTACHMENT_CHECKPOINTING_V1 UINT32_C(1)

#define POTLUCKDB_RUNTIME_REBASE_NONE UINT32_C(0)
#define POTLUCKDB_RUNTIME_REBASE_STAGED UINT32_C(1)
#define POTLUCKDB_RUNTIME_REBASE_REPLAYING UINT32_C(2)

#define POTLUCKDB_RUNTIME_SYNC_PUBLISHED_V1 UINT32_C(1)
#define POTLUCKDB_RUNTIME_SYNC_CONFLICT_V1 UINT32_C(2)
#define POTLUCKDB_RUNTIME_SYNC_OPERATION_EXCHANGE_V1 UINT32_C(1)
#define POTLUCKDB_RUNTIME_SYNC_OPERATION_COMPLETION_V1 UINT32_C(2)
#define POTLUCKDB_RUNTIME_SYNC_OPERATION_COMPLETE_V1 UINT32_C(3)
#define POTLUCKDB_RUNTIME_ALL_PENDING_V1 UINT64_MAX

#define POTLUCKDB_RUNTIME_SYNC_AUTOMATIC_V1 UINT32_C(1)
#define POTLUCKDB_RUNTIME_SYNC_MANUAL_V1 UINT32_C(2)
#define POTLUCKDB_RUNTIME_AUTO_SYNC_MANUAL_V1 UINT32_C(1)
#define POTLUCKDB_RUNTIME_AUTO_SYNC_UNBOUND_V1 UINT32_C(2)
#define POTLUCKDB_RUNTIME_AUTO_SYNC_SCHEDULED_V1 UINT32_C(3)
#define POTLUCKDB_RUNTIME_AUTO_SYNC_SYNCHRONIZING_V1 UINT32_C(4)
#define POTLUCKDB_RUNTIME_AUTO_SYNC_IDLE_V1 UINT32_C(5)
#define POTLUCKDB_RUNTIME_AUTO_SYNC_BACKOFF_V1 UINT32_C(6)
#define POTLUCKDB_RUNTIME_AUTO_SYNC_CONFLICT_V1 UINT32_C(7)
#define POTLUCKDB_RUNTIME_AUTO_SYNC_FAILED_V1 UINT32_C(8)
#define POTLUCKDB_RUNTIME_AUTO_SYNC_CLOSED_V1 UINT32_C(9)

#define POTLUCKDB_RUNTIME_TRANSACTION_DEFERRED_V1 UINT32_C(1)
#define POTLUCKDB_RUNTIME_TRANSACTION_IMMEDIATE_V1 UINT32_C(2)
#define POTLUCKDB_RUNTIME_TRANSACTION_EXCLUSIVE_V1 UINT32_C(3)

#define POTLUCKDB_RUNTIME_LOCAL_PUT_V1 UINT32_C(1)
#define POTLUCKDB_RUNTIME_LOCAL_APPEND_V1 UINT32_C(2)
#define POTLUCKDB_RUNTIME_LOCAL_DELETE_V1 UINT32_C(3)

typedef uint64_t PotluckDbRuntimeHandleV1;

typedef struct PotluckDbRuntimeSliceV1 {
    const uint8_t *data;
    uint64_t length;
} PotluckDbRuntimeSliceV1;

typedef struct PotluckDbRuntimeHostErrorV1 {
    uint32_t struct_size;
    uint32_t abi_version;
    int32_t sqlite_primary_code;
    int32_t sqlite_extended_code;
    PotluckDbRuntimeSliceV1 message;
} PotluckDbRuntimeHostErrorV1;

typedef struct PotluckDbRuntimeBufferV1 {
    uint64_t token;
    const uint8_t *data;
    uint64_t length;
} PotluckDbRuntimeBufferV1;

/* Buffer data is immutable and valid until its nonzero token is released once. */
/* Every nonzero handle and resource token shares one nonreused namespace per loaded interface. */

typedef struct PotluckDbRuntimeErrorV1 {
    uint32_t struct_size;
    uint32_t abi_version;
    int32_t code;
    int32_t sqlite_primary_code;
    int32_t sqlite_extended_code;
    uint32_t reserved;
    PotluckDbRuntimeBufferV1 message;
} PotluckDbRuntimeErrorV1;

typedef struct PotluckDbRuntimeOpenOptionsV1 {
    uint32_t struct_size;
    uint32_t abi_version;
    PotluckDbRuntimeSliceV1 root;
    PotluckDbRuntimeSliceV1 client_id;
    PotluckDbRuntimeSliceV1 database_path;
    uint64_t budget_bytes;
    uint64_t minimum_budget_bytes;
    uint64_t quarantine_limit;
    uint64_t local_compact_threshold_bytes;
    uint32_t engine_profile;
    uint32_t flags;
    uint8_t encryption_key[32];
} PotluckDbRuntimeOpenOptionsV1;

typedef struct PotluckDbRuntimeAttachmentOptionsV1 {
    uint32_t struct_size;
    uint32_t abi_version;
    uint32_t wal_validation;
    uint32_t flags;
    uint64_t checkpoint_soft_limit_bytes;
    uint64_t checkpoint_forced_threshold_bytes;
} PotluckDbRuntimeAttachmentOptionsV1;

typedef struct PotluckDbRuntimeSparseOptionsV1 {
    uint32_t struct_size;
    uint32_t abi_version;
    uint32_t flags;
    uint32_t reserved;
    uint8_t revision[32];
} PotluckDbRuntimeSparseOptionsV1;

typedef struct PotluckDbRuntimePrefetchOptionsV1 {
    uint32_t struct_size;
    uint32_t abi_version;
    uint32_t policy_kind;
    uint32_t flags;
    uint64_t background_target_bytes;
    uint64_t memory_budget_bytes;
    uint32_t absolute_max_concurrency;
    uint32_t maximum_pending_pages;
    uint32_t fixed_limit;
    uint32_t adaptive_probe;
    uint32_t adaptive_initial;
    uint32_t adaptive_maximum;
    uint32_t adaptive_sample_size;
    uint32_t adaptive_starting;
    uint32_t statement_sequential_lookahead;
    uint32_t statement_maximum_traces;
    uint32_t statement_maximum_trace_pages;
    uint32_t reserved;
} PotluckDbRuntimePrefetchOptionsV1;

typedef struct PotluckDbRuntimePrefetchStatusV1 {
    uint32_t struct_size;
    uint32_t abi_version;
    uint32_t enabled;
    uint32_t background_state;
    uint64_t limit;
    uint64_t active;
    uint64_t queued;
    uint64_t ready_pages;
    uint64_t ready_bytes;
    uint64_t background_target_bytes;
    uint64_t demand_submitted;
    uint64_t demand_joined;
    uint64_t demand_ready;
    uint64_t statement_submitted;
    uint64_t background_submitted;
    uint64_t speculative_cancelled;
    uint64_t speculative_completed;
    uint64_t speculative_discarded;
    uint64_t speculative_failed;
    uint64_t background_installed;
    uint64_t background_install_refused;
    uint64_t background_batches;
    uint64_t background_install_failed;
    uint64_t policy_failures;
    uint64_t limit_increases;
    uint64_t limit_decreases;
} PotluckDbRuntimePrefetchStatusV1;

typedef struct PotluckDbRuntimeSessionStatusV1 {
    uint32_t struct_size;
    uint32_t abi_version;
    uint32_t sqlite_attached;
    uint32_t read_only;
    uint32_t closed;
    uint32_t reserved;
} PotluckDbRuntimeSessionStatusV1;

typedef struct PotluckDbRuntimeSyncPolicyV1 {
    uint32_t struct_size;
    uint32_t abi_version;
    uint32_t mode;
    uint32_t reserved;
    uint64_t outbound_debounce_ms;
    uint64_t active_pull_interval_ms;
    uint64_t idle_pull_interval_ms;
    uint64_t retry_initial_ms;
    uint64_t retry_max_ms;
} PotluckDbRuntimeSyncPolicyV1;

typedef struct PotluckDbRuntimeAutoSyncStatusV1 {
    uint32_t struct_size;
    uint32_t abi_version;
    uint32_t mode;
    uint32_t state;
    uint32_t has_next_delay;
    uint32_t has_conflict_remote_revision;
    uint32_t consecutive_failures;
    uint32_t reserved;
    uint64_t next_delay_ms;
    uint64_t outbound_debounce_ms;
    uint64_t active_pull_interval_ms;
    uint64_t idle_pull_interval_ms;
    uint64_t retry_initial_ms;
    uint64_t retry_max_ms;
    uint64_t attempts;
    uint64_t publications;
    uint64_t pulls;
    uint64_t unchanged;
    uint64_t conflicts;
    uint64_t transient_failures;
    uint64_t permanent_failures;
    int32_t last_error_code;
    uint32_t reserved_error;
    uint8_t conflict_remote_revision[32];
    PotluckDbRuntimeBufferV1 last_error;
    PotluckDbRuntimeBufferV1 conflict_reason;
} PotluckDbRuntimeAutoSyncStatusV1;

typedef struct PotluckDbRuntimeBranchStatusV1 {
    uint32_t struct_size;
    uint32_t abi_version;
    uint32_t engine_profile;
    uint32_t rebase_phase;
    uint64_t piece_count;
    uint64_t pending;
    uint64_t resident_bytes;
    uint64_t protected_bytes;
    uint64_t budget_bytes;
    uint64_t replay_pending;
    uint64_t quarantine_limit;
    uint64_t quarantine_tombstones;
    uint64_t quarantined_count;
    uint32_t has_base_revision;
    uint32_t has_frozen_sync;
    uint8_t base_revision[32];
    uint8_t frozen_sync[32];
} PotluckDbRuntimeBranchStatusV1;

typedef struct PotluckDbRuntimeCheckpointResultV1 {
    uint32_t struct_size;
    uint32_t abi_version;
    int32_t log_frames;
    int32_t checkpointed_frames;
} PotluckDbRuntimeCheckpointResultV1;

typedef struct PotluckDbRuntimeCheckpointStatusV1 {
    uint32_t struct_size;
    uint32_t abi_version;
    uint32_t available;
    uint32_t due;
    uint32_t scheduled;
    uint32_t reserved;
    uint64_t soft_limit_bytes;
    uint64_t forced_checkpoint_threshold_bytes;
    uint64_t wal_bytes;
    uint64_t soft_scheduled;
    uint64_t forced_checkpoint_attempts;
    uint64_t manual;
    uint64_t completed;
    uint64_t failures;
    uint64_t stale_commands;
    uint64_t pages_verified;
    uint64_t wal_bytes_reclaimed;
} PotluckDbRuntimeCheckpointStatusV1;

typedef struct PotluckDbRuntimeExchangeOptionsV1 {
    uint32_t struct_size;
    uint32_t abi_version;
    uint32_t immutable_read_concurrency;
    uint32_t immutable_write_concurrency;
    uint64_t retention_target_revisions;
    uint64_t retention_high_watermark;
} PotluckDbRuntimeExchangeOptionsV1;

typedef struct PotluckDbRuntimeTransferObservationV1 {
    uint32_t struct_size;
    uint32_t abi_version;
    uint64_t service_ns;
    uint32_t backlogged;
    uint32_t outcome;
} PotluckDbRuntimeTransferObservationV1;

typedef int32_t (*PotluckDbRuntimeConcurrencyStartV1)(
    void *context,
    uint64_t *out_limit,
    PotluckDbRuntimeHostErrorV1 *error
);
typedef int32_t (*PotluckDbRuntimeConcurrencyObserveV1)(
    void *context,
    PotluckDbRuntimeTransferObservationV1 observation,
    uint64_t *out_limit,
    PotluckDbRuntimeHostErrorV1 *error
);
typedef int32_t (*PotluckDbRuntimeConcurrencyResetV1)(
    void *context,
    uint64_t *out_limit,
    PotluckDbRuntimeHostErrorV1 *error
);

typedef struct PotluckDbRuntimeConcurrencyPolicyV1 {
    uint32_t struct_size;
    uint32_t abi_version;
    void *context;
    PotluckDbRuntimeConcurrencyStartV1 start;
    PotluckDbRuntimeConcurrencyObserveV1 observe;
    PotluckDbRuntimeConcurrencyResetV1 reset;
} PotluckDbRuntimeConcurrencyPolicyV1;

typedef struct PotluckDbRuntimeUploadOptionsV1 {
    uint32_t struct_size;
    uint32_t abi_version;
    uint32_t policy_kind;
    uint32_t flags;
    uint32_t absolute_max_concurrency;
    uint32_t maximum_pending_objects;
    uint32_t fixed_limit;
    uint32_t adaptive_probe;
    uint32_t adaptive_initial;
    uint32_t adaptive_maximum;
    uint32_t adaptive_sample_size;
    uint32_t adaptive_starting;
    const PotluckDbRuntimeConcurrencyPolicyV1 *custom_policy;
} PotluckDbRuntimeUploadOptionsV1;

typedef struct PotluckDbRuntimeKeyRotationOptionsV1 {
    uint32_t struct_size;
    uint32_t abi_version;
    uint32_t engine_profile;
    uint32_t flags;
    uint32_t immutable_read_concurrency;
    uint32_t immutable_write_concurrency;
    PotluckDbRuntimeSliceV1 client_id;
    uint8_t source_key[32];
    uint8_t new_key[32];
    uint8_t revision[32];
} PotluckDbRuntimeKeyRotationOptionsV1;

typedef struct PotluckDbRuntimeKeyRotationReportV1 {
    uint32_t struct_size;
    uint32_t abi_version;
    uint32_t has_source_key_id;
    uint32_t published;
    uint64_t piece_count;
    uint8_t source_revision[32];
    uint8_t target_revision[32];
    uint8_t source_key_id[32];
    uint8_t target_key_id[32];
} PotluckDbRuntimeKeyRotationReportV1;

typedef struct PotluckDbRuntimeCollectionReportV1 {
    uint32_t struct_size;
    uint32_t abi_version;
    uint32_t available;
    uint32_t aborted;
    uint32_t has_floor_before;
    uint32_t has_floor_after;
    uint8_t floor_before[32];
    uint8_t floor_after[32];
    uint64_t epoch;
    uint64_t marked;
    uint64_t swept;
} PotluckDbRuntimeCollectionReportV1;

typedef struct PotluckDbRuntimeExchangeStatusV1 {
    uint32_t struct_size;
    uint32_t abi_version;
    uint32_t has_retained_revisions;
    uint32_t maintenance_pending;
    uint64_t retained_revisions;
    uint64_t policy_failures;
    uint32_t has_last_collection;
    uint32_t last_collection_aborted;
    uint32_t has_floor_before;
    uint32_t has_floor_after;
    uint8_t floor_before[32];
    uint8_t floor_after[32];
    uint64_t last_collection_epoch;
    uint64_t last_collection_marked;
    uint64_t last_collection_swept;
} PotluckDbRuntimeExchangeStatusV1;

typedef struct PotluckDbRuntimeSyncOutcomeV1 {
    uint32_t struct_size;
    uint32_t abi_version;
    uint32_t kind;
    uint32_t reserved;
    uint64_t token;
    uint8_t identifier[32];
    uint32_t has_base_revision;
    uint32_t has_remote_revision;
    uint32_t has_remote_epoch;
    uint32_t reserved_flags;
    uint8_t base_revision[32];
    uint8_t revision[32];
    uint8_t remote_revision[32];
    uint64_t head_epoch;
    uint64_t remote_epoch;
    uint64_t transaction_count;
    PotluckDbRuntimeBufferV1 reason;
} PotluckDbRuntimeSyncOutcomeV1;

typedef struct PotluckDbRuntimeSyncOperationStatusV1 {
    uint32_t struct_size;
    uint32_t abi_version;
    uint32_t phase;
    uint32_t has_prepared;
    uint64_t prepared_token;
    uint64_t outcome_token;
} PotluckDbRuntimeSyncOperationStatusV1;

typedef struct PotluckDbRuntimeRebaseResultV1 {
    uint32_t struct_size;
    uint32_t abi_version;
    uint32_t connection_reset;
    uint32_t reserved;
    uint8_t remote_revision[32];
    uint64_t already_published_count;
    uint64_t replayed_count;
    uint64_t quarantined_count;
    uint64_t pending_count;
    PotluckDbRuntimeBufferV1 details;
} PotluckDbRuntimeRebaseResultV1;

typedef struct PotluckDbRuntimeStatementCallV1 {
    uint32_t struct_size;
    uint32_t abi_version;
    PotluckDbRuntimeSliceV1 sql;
    PotluckDbRuntimeSliceV1 bindings;
} PotluckDbRuntimeStatementCallV1;

typedef struct PotluckDbRuntimeDuckDbOptionsV1 {
    uint32_t struct_size;
    uint32_t abi_version;
    PotluckDbRuntimeSliceV1 source;
    uint32_t reader_count;
    uint32_t flags;
} PotluckDbRuntimeDuckDbOptionsV1;

typedef struct PotluckDbRuntimeDuckDbQueryMetricsV1 {
    uint32_t struct_size;
    uint32_t abi_version;
    uint32_t flags;
    uint32_t reserved;
    uint64_t source_tables;
    uint64_t partitions;
    uint64_t readers_used;
    uint64_t batches;
    uint64_t scanned_rows;
    uint64_t maximum_callbacks_in_flight;
} PotluckDbRuntimeDuckDbQueryMetricsV1;

/* details is one big-endian SVQ1 value in this order: already-published string list,
 * replayed string list, quarantine list, and pending string list. Every list starts
 * with a uint64 count; every UTF-8 string or opaque byte value starts with a uint64
 * length. A quarantine is sequence uint64, has-error-code byte, error-code int32,
 * error name, error message, recipe transaction ID, statement-count uint64, then
 * SQL and canonical binding bytes for each statement. Release details once. */

typedef struct PotluckDbRuntimeLocalOperationV1 {
    uint32_t kind;
    uint32_t reserved;
    PotluckDbRuntimeSliceV1 name;
    PotluckDbRuntimeSliceV1 data;
} PotluckDbRuntimeLocalOperationV1;

typedef struct PotluckDbRuntimeLocalCounterSnapshotV1 {
    uint32_t struct_size;
    uint32_t abi_version;
    uint64_t reads;
    uint64_t lists;
    uint64_t commit_calls;
    uint64_t groups_applied;
    uint64_t groups_retried;
    uint64_t operations_applied;
    uint64_t value_bytes_written;
    uint64_t bytes_read;
    uint64_t flushes;
    uint64_t tail_truncations;
    uint64_t compactions;
    uint64_t compacted_bytes;
} PotluckDbRuntimeLocalCounterSnapshotV1;

typedef struct PotluckDbRuntimeBucketCounterSnapshotV1 {
    uint32_t struct_size;
    uint32_t abi_version;
    uint64_t gets;
    uint64_t puts;
    uint64_t idempotent_puts;
    uint64_t exists;
    uint64_t lists;
    uint64_t deletes;
    uint64_t cell_gets;
    uint64_t cell_swaps;
    uint64_t cell_swap_wins;
    uint64_t transport_gets;
    uint64_t transport_puts;
    uint64_t transport_heads;
    uint64_t transport_lists;
    uint64_t transport_deletes;
    uint64_t transport_bytes_received;
    uint64_t transport_bytes_submitted;
} PotluckDbRuntimeBucketCounterSnapshotV1;

typedef int32_t (*PotluckDbRuntimeEmitBytesV1)(void *context, PotluckDbRuntimeSliceV1 value);
typedef int32_t (*PotluckDbRuntimeLocalReadV1)(
    void *context,
    PotluckDbRuntimeSliceV1 name,
    PotluckDbRuntimeEmitBytesV1 emit,
    void *emit_context,
    PotluckDbRuntimeHostErrorV1 *error
);
typedef int32_t (*PotluckDbRuntimeLocalListV1)(
    void *context,
    PotluckDbRuntimeSliceV1 prefix,
    PotluckDbRuntimeEmitBytesV1 emit,
    void *emit_context,
    PotluckDbRuntimeHostErrorV1 *error
);
typedef int32_t (*PotluckDbRuntimeLocalCommitV1)(
    void *context,
    PotluckDbRuntimeSliceV1 group_id,
    const PotluckDbRuntimeLocalOperationV1 *operations,
    uint64_t operation_count,
    uint32_t *out_applied,
    PotluckDbRuntimeHostErrorV1 *error
);
typedef int32_t (*PotluckDbRuntimeLocalCloseV1)(
    void *context,
    PotluckDbRuntimeHostErrorV1 *error
);
typedef void (*PotluckDbRuntimeLocalCountersV1)(
    void *context,
    PotluckDbRuntimeLocalCounterSnapshotV1 *out_counters
);
typedef void (*PotluckDbRuntimeLocalDestroyV1)(void *context);

typedef struct PotluckDbRuntimeLocalAdapterV1 {
    uint32_t struct_size;
    uint32_t abi_version;
    void *context;
    PotluckDbRuntimeLocalReadV1 read;
    PotluckDbRuntimeLocalListV1 list;
    PotluckDbRuntimeLocalCommitV1 commit;
    PotluckDbRuntimeLocalCloseV1 close;
    PotluckDbRuntimeLocalCountersV1 counters;
    PotluckDbRuntimeLocalDestroyV1 destroy;
} PotluckDbRuntimeLocalAdapterV1;

typedef int32_t (*PotluckDbRuntimeBucketGetV1)(
    void *context,
    PotluckDbRuntimeSliceV1 key,
    PotluckDbRuntimeEmitBytesV1 emit,
    void *emit_context,
    PotluckDbRuntimeHostErrorV1 *error
);
typedef int32_t (*PotluckDbRuntimeBucketPutV1)(
    void *context,
    PotluckDbRuntimeSliceV1 key,
    PotluckDbRuntimeSliceV1 data,
    PotluckDbRuntimeHostErrorV1 *error
);
typedef int32_t (*PotluckDbRuntimeBucketExistsV1)(
    void *context,
    PotluckDbRuntimeSliceV1 key,
    uint32_t *out_exists,
    PotluckDbRuntimeHostErrorV1 *error
);
typedef int32_t (*PotluckDbRuntimeBucketListV1)(
    void *context,
    PotluckDbRuntimeSliceV1 prefix,
    PotluckDbRuntimeEmitBytesV1 emit,
    void *emit_context,
    PotluckDbRuntimeHostErrorV1 *error
);
typedef int32_t (*PotluckDbRuntimeBucketDeleteV1)(
    void *context,
    PotluckDbRuntimeSliceV1 key,
    PotluckDbRuntimeHostErrorV1 *error
);
typedef int32_t (*PotluckDbRuntimeBucketCellGetV1)(
    void *context,
    PotluckDbRuntimeEmitBytesV1 emit,
    void *emit_context,
    PotluckDbRuntimeHostErrorV1 *error
);
typedef int32_t (*PotluckDbRuntimeBucketCellSwapV1)(
    void *context,
    uint32_t expected_present,
    PotluckDbRuntimeSliceV1 expected,
    PotluckDbRuntimeSliceV1 new_value,
    uint32_t *out_swapped,
    PotluckDbRuntimeHostErrorV1 *error
);
typedef int32_t (*PotluckDbRuntimeBucketCellPairGetV1)(
    void *context,
    PotluckDbRuntimeEmitBytesV1 emit_head,
    void *head_context,
    PotluckDbRuntimeEmitBytesV1 emit_floor,
    void *floor_context,
    PotluckDbRuntimeHostErrorV1 *error
);
typedef void (*PotluckDbRuntimeBucketCountersV1)(
    void *context,
    PotluckDbRuntimeBucketCounterSnapshotV1 *out_counters
);

typedef struct PotluckDbRuntimeBucketAdapterV1 {
    uint32_t struct_size;
    uint32_t abi_version;
    void *context;
    PotluckDbRuntimeBucketGetV1 get;
    PotluckDbRuntimeBucketPutV1 put;
    PotluckDbRuntimeBucketExistsV1 exists;
    PotluckDbRuntimeBucketListV1 list;
    PotluckDbRuntimeBucketDeleteV1 delete_value;
    PotluckDbRuntimeBucketCellGetV1 head_get;
    PotluckDbRuntimeBucketCellSwapV1 head_swap;
    PotluckDbRuntimeBucketCellGetV1 floor_get;
    PotluckDbRuntimeBucketCellSwapV1 floor_swap;
    PotluckDbRuntimeBucketCellPairGetV1 cell_pair_get;
    PotluckDbRuntimeBucketCountersV1 counters;
} PotluckDbRuntimeBucketAdapterV1;

/* Input and output regions passed to one call must not overlap. */
/* Before reusing an error, release its nonzero message token exactly once. */
/* Host callbacks must not unwind; emitted values are copied during the callback. */
/* A host-error message must remain readable until the next adapter callback begins. */
/* Bucket callbacks may overlap and must be thread-safe; immutable GET/PUT calls are bounded. */
/* A bucket host-error message must remain immutable through the enclosing ABI call. */
/* Bucket adapter and context ownership always remain with the caller. */
/* Concurrency policy callbacks run on the exchange caller and must not unwind. */
/* Upload options and custom policy storage remain caller-owned until the call returns. */

uint32_t potluckdb_runtime_abi_version_v1(void);
uint32_t potluckdb_runtime_open_options_size_v1(void);
uint32_t potluckdb_runtime_attachment_options_size_v1(void);
uint32_t potluckdb_runtime_sparse_options_size_v1(void);
uint32_t potluckdb_runtime_prefetch_options_size_v1(void);
uint32_t potluckdb_runtime_prefetch_status_size_v1(void);
uint32_t potluckdb_runtime_error_size_v1(void);
uint32_t potluckdb_runtime_branch_status_size_v1(void);
uint32_t potluckdb_runtime_session_status_size_v1(void);
uint32_t potluckdb_runtime_sync_policy_size_v1(void);
uint32_t potluckdb_runtime_auto_sync_status_size_v1(void);
uint32_t potluckdb_runtime_checkpoint_result_size_v1(void);
uint32_t potluckdb_runtime_checkpoint_status_size_v1(void);
uint32_t potluckdb_runtime_host_error_size_v1(void);
uint32_t potluckdb_runtime_local_operation_size_v1(void);
uint32_t potluckdb_runtime_local_counter_snapshot_size_v1(void);
uint32_t potluckdb_runtime_local_adapter_size_v1(void);
uint32_t potluckdb_runtime_bucket_counter_snapshot_size_v1(void);
uint32_t potluckdb_runtime_bucket_adapter_size_v1(void);
uint32_t potluckdb_runtime_exchange_options_size_v1(void);
uint32_t potluckdb_runtime_transfer_observation_size_v1(void);
uint32_t potluckdb_runtime_concurrency_policy_size_v1(void);
uint32_t potluckdb_runtime_upload_options_size_v1(void);
uint32_t potluckdb_runtime_key_rotation_options_size_v1(void);
uint32_t potluckdb_runtime_key_rotation_report_size_v1(void);
uint32_t potluckdb_runtime_collection_report_size_v1(void);
uint32_t potluckdb_runtime_exchange_status_size_v1(void);
uint32_t potluckdb_runtime_sync_outcome_size_v1(void);
uint32_t potluckdb_runtime_sync_operation_status_size_v1(void);
uint32_t potluckdb_runtime_rebase_result_size_v1(void);
uint32_t potluckdb_runtime_statement_call_size_v1(void);
uint32_t potluckdb_runtime_duckdb_options_size_v1(void);
uint32_t potluckdb_runtime_duckdb_query_metrics_size_v1(void);

void potluckdb_runtime_open_options_init_v1(PotluckDbRuntimeOpenOptionsV1 *options);
void potluckdb_runtime_attachment_options_init_v1(PotluckDbRuntimeAttachmentOptionsV1 *options);
void potluckdb_runtime_sparse_options_init_v1(PotluckDbRuntimeSparseOptionsV1 *options);
void potluckdb_runtime_prefetch_options_init_v1(PotluckDbRuntimePrefetchOptionsV1 *options);
void potluckdb_runtime_prefetch_status_init_v1(PotluckDbRuntimePrefetchStatusV1 *status);
void potluckdb_runtime_error_init_v1(PotluckDbRuntimeErrorV1 *error);
void potluckdb_runtime_branch_status_init_v1(PotluckDbRuntimeBranchStatusV1 *status);
void potluckdb_runtime_session_status_init_v1(PotluckDbRuntimeSessionStatusV1 *status);
void potluckdb_runtime_sync_policy_init_v1(PotluckDbRuntimeSyncPolicyV1 *policy);
void potluckdb_runtime_auto_sync_status_init_v1(PotluckDbRuntimeAutoSyncStatusV1 *status);
void potluckdb_runtime_checkpoint_result_init_v1(PotluckDbRuntimeCheckpointResultV1 *result);
void potluckdb_runtime_checkpoint_status_init_v1(PotluckDbRuntimeCheckpointStatusV1 *status);
void potluckdb_runtime_host_error_init_v1(PotluckDbRuntimeHostErrorV1 *error);
void potluckdb_runtime_local_counter_snapshot_init_v1(PotluckDbRuntimeLocalCounterSnapshotV1 *counters);
void potluckdb_runtime_local_adapter_init_v1(PotluckDbRuntimeLocalAdapterV1 *adapter);
void potluckdb_runtime_bucket_counter_snapshot_init_v1(PotluckDbRuntimeBucketCounterSnapshotV1 *counters);
void potluckdb_runtime_bucket_adapter_init_v1(PotluckDbRuntimeBucketAdapterV1 *adapter);
void potluckdb_runtime_exchange_options_init_v1(PotluckDbRuntimeExchangeOptionsV1 *options);
void potluckdb_runtime_transfer_observation_init_v1(PotluckDbRuntimeTransferObservationV1 *observation);
void potluckdb_runtime_concurrency_policy_init_v1(PotluckDbRuntimeConcurrencyPolicyV1 *policy);
void potluckdb_runtime_upload_options_init_v1(PotluckDbRuntimeUploadOptionsV1 *options);
void potluckdb_runtime_key_rotation_options_init_v1(PotluckDbRuntimeKeyRotationOptionsV1 *options);
void potluckdb_runtime_key_rotation_report_init_v1(PotluckDbRuntimeKeyRotationReportV1 *report);
void potluckdb_runtime_collection_report_init_v1(PotluckDbRuntimeCollectionReportV1 *report);
void potluckdb_runtime_exchange_status_init_v1(PotluckDbRuntimeExchangeStatusV1 *status);
void potluckdb_runtime_sync_outcome_init_v1(PotluckDbRuntimeSyncOutcomeV1 *outcome);
void potluckdb_runtime_sync_operation_status_init_v1(PotluckDbRuntimeSyncOperationStatusV1 *status);
void potluckdb_runtime_rebase_result_init_v1(PotluckDbRuntimeRebaseResultV1 *result);
void potluckdb_runtime_statement_call_init_v1(PotluckDbRuntimeStatementCallV1 *call);
void potluckdb_runtime_duckdb_options_init_v1(PotluckDbRuntimeDuckDbOptionsV1 *options);
void potluckdb_runtime_duckdb_query_metrics_init_v1(PotluckDbRuntimeDuckDbQueryMetricsV1 *metrics);

int32_t potluckdb_runtime_session_open_v1(
    const PotluckDbRuntimeOpenOptionsV1 *options,
    PotluckDbRuntimeHandleV1 *out_handle,
    PotluckDbRuntimeErrorV1 *out_error
);
/* Configured open keeps session and SQLite-attachment policy in separate V1 records. */
int32_t potluckdb_runtime_session_open_configured_v1(
    const PotluckDbRuntimeOpenOptionsV1 *options,
    const PotluckDbRuntimeAttachmentOptionsV1 *attachment,
    PotluckDbRuntimeHandleV1 *out_handle,
    PotluckDbRuntimeErrorV1 *out_error
);
/* On ownership transfer, adapter is zeroed; close then destroy each run exactly once. */
int32_t potluckdb_runtime_session_open_with_local_v1(
    const PotluckDbRuntimeOpenOptionsV1 *options,
    PotluckDbRuntimeLocalAdapterV1 *adapter,
    PotluckDbRuntimeHandleV1 *out_handle,
    PotluckDbRuntimeErrorV1 *out_error
);
/* Validation of both option records precedes local-adapter ownership transfer. */
int32_t potluckdb_runtime_session_open_with_local_configured_v1(
    const PotluckDbRuntimeOpenOptionsV1 *options,
    const PotluckDbRuntimeAttachmentOptionsV1 *attachment,
    PotluckDbRuntimeLocalAdapterV1 *adapter,
    PotluckDbRuntimeHandleV1 *out_handle,
    PotluckDbRuntimeErrorV1 *out_error
);
/* The exchange stays caller-owned and cannot close while the sparse session is attached. */
int32_t potluckdb_runtime_session_open_sparse_v1(
    const PotluckDbRuntimeOpenOptionsV1 *options,
    const PotluckDbRuntimeAttachmentOptionsV1 *attachment,
    const PotluckDbRuntimeSparseOptionsV1 *sparse,
    PotluckDbRuntimeHandleV1 exchange_handle,
    PotluckDbRuntimeHandleV1 *out_handle,
    PotluckDbRuntimeErrorV1 *out_error
);
/* Additive opt-in: all selection, decoding, installation, and counters remain native-owned. */
int32_t potluckdb_runtime_session_open_sparse_prefetch_v1(
    const PotluckDbRuntimeOpenOptionsV1 *options,
    const PotluckDbRuntimeAttachmentOptionsV1 *attachment,
    const PotluckDbRuntimeSparseOptionsV1 *sparse,
    const PotluckDbRuntimePrefetchOptionsV1 *prefetch,
    PotluckDbRuntimeHandleV1 exchange_handle,
    PotluckDbRuntimeHandleV1 *out_handle,
    PotluckDbRuntimeErrorV1 *out_error
);
/* Validation precedes local-adapter transfer; exact-revision mode is read-only. */
int32_t potluckdb_runtime_session_open_with_local_sparse_v1(
    const PotluckDbRuntimeOpenOptionsV1 *options,
    const PotluckDbRuntimeAttachmentOptionsV1 *attachment,
    const PotluckDbRuntimeSparseOptionsV1 *sparse,
    PotluckDbRuntimeLocalAdapterV1 *adapter,
    PotluckDbRuntimeHandleV1 exchange_handle,
    PotluckDbRuntimeHandleV1 *out_handle,
    PotluckDbRuntimeErrorV1 *out_error
);
/* Validation precedes local-adapter transfer; the owner still performs every local mutation. */
int32_t potluckdb_runtime_session_open_with_local_sparse_prefetch_v1(
    const PotluckDbRuntimeOpenOptionsV1 *options,
    const PotluckDbRuntimeAttachmentOptionsV1 *attachment,
    const PotluckDbRuntimeSparseOptionsV1 *sparse,
    const PotluckDbRuntimePrefetchOptionsV1 *prefetch,
    PotluckDbRuntimeLocalAdapterV1 *adapter,
    PotluckDbRuntimeHandleV1 exchange_handle,
    PotluckDbRuntimeHandleV1 *out_handle,
    PotluckDbRuntimeErrorV1 *out_error
);
/* Each returned handle owns one independent read/write SQLite connection. */
/* A null prefetch pointer disables prefetch for this connection. */
int32_t potluckdb_runtime_session_connect_v1(
    PotluckDbRuntimeHandleV1 session_handle,
    const PotluckDbRuntimeAttachmentOptionsV1 *attachment,
    const PotluckDbRuntimePrefetchOptionsV1 *prefetch,
    PotluckDbRuntimeHandleV1 *out_connection,
    PotluckDbRuntimeErrorV1 *out_error
);
int32_t potluckdb_runtime_session_connection_count_v1(
    PotluckDbRuntimeHandleV1 session_handle,
    uint64_t *out_count,
    PotluckDbRuntimeErrorV1 *out_error
);
/* Policy decisions are runtime-owned; the host only supplies monotonic elapsed time and wakeups. */
int32_t potluckdb_runtime_session_set_sync_policy_v1(
    PotluckDbRuntimeHandleV1 session_handle,
    const PotluckDbRuntimeSyncPolicyV1 *policy,
    PotluckDbRuntimeErrorV1 *out_error
);
/* Status owns last_error and conflict_reason buffers; release each nonzero token once. */
int32_t potluckdb_runtime_session_auto_sync_status_v1(
    PotluckDbRuntimeHandleV1 session_handle,
    PotluckDbRuntimeAutoSyncStatusV1 *out_status,
    uint32_t out_status_size,
    PotluckDbRuntimeErrorV1 *out_error
);
/* Schedule this blocking driver call off an event loop, then wait next_delay_ms before driving. */
int32_t potluckdb_runtime_session_drive_auto_sync_v1(
    PotluckDbRuntimeHandleV1 session_handle,
    uint64_t elapsed_ms,
    PotluckDbRuntimeAutoSyncStatusV1 *out_status,
    uint32_t out_status_size,
    PotluckDbRuntimeErrorV1 *out_error
);
/* The source is a shared-library path for the native-C platform binding. */
/* Other platform bindings interpret it as their DuckDB module source. */
int32_t potluckdb_runtime_session_open_duckdb_v1(
    PotluckDbRuntimeHandleV1 session_handle,
    const PotluckDbRuntimeDuckDbOptionsV1 *options,
    PotluckDbRuntimeHandleV1 *out_handle,
    PotluckDbRuntimeErrorV1 *out_error
);
/* DQR1 is an owned typed result with 128-bit integers; release its buffer token once. */
int32_t potluckdb_runtime_duckdb_query_v1(
    PotluckDbRuntimeHandleV1 handle,
    PotluckDbRuntimeSliceV1 sql,
    PotluckDbRuntimeBufferV1 *out_rows,
    PotluckDbRuntimeDuckDbQueryMetricsV1 *out_metrics,
    uint32_t out_metrics_size,
    PotluckDbRuntimeErrorV1 *out_error
);
int32_t potluckdb_runtime_duckdb_library_version_v1(
    PotluckDbRuntimeHandleV1 handle,
    PotluckDbRuntimeBufferV1 *out_version,
    PotluckDbRuntimeErrorV1 *out_error
);
int32_t potluckdb_runtime_duckdb_reader_count_v1(
    PotluckDbRuntimeHandleV1 handle,
    uint32_t *out_count,
    PotluckDbRuntimeErrorV1 *out_error
);
int32_t potluckdb_runtime_duckdb_close_v1(
    PotluckDbRuntimeHandleV1 handle,
    PotluckDbRuntimeErrorV1 *out_error
);
int32_t potluckdb_runtime_duckdb_release_v1(
    PotluckDbRuntimeHandleV1 handle,
    PotluckDbRuntimeErrorV1 *out_error
);
int32_t potluckdb_runtime_exchange_open_filesystem_v1(
    const PotluckDbRuntimeExchangeOptionsV1 *options,
    PotluckDbRuntimeSliceV1 root,
    PotluckDbRuntimeHandleV1 *out_handle,
    PotluckDbRuntimeErrorV1 *out_error
);
/* The adapter table is copied; its caller-owned context remains live until release. */
int32_t potluckdb_runtime_exchange_open_with_bucket_v1(
    const PotluckDbRuntimeExchangeOptionsV1 *options,
    const PotluckDbRuntimeBucketAdapterV1 *adapter,
    PotluckDbRuntimeHandleV1 *out_handle,
    PotluckDbRuntimeErrorV1 *out_error
);
/* Both distinct exchange handles remain open and are held exclusively until every worker exits. */
int32_t potluckdb_runtime_rotate_encryption_key_v1(
    const PotluckDbRuntimeKeyRotationOptionsV1 *options,
    PotluckDbRuntimeHandleV1 source_exchange_handle,
    PotluckDbRuntimeHandleV1 target_exchange_handle,
    PotluckDbRuntimeKeyRotationReportV1 *out_report,
    uint32_t out_report_size,
    PotluckDbRuntimeErrorV1 *out_error
);
/* A zero prepared token means the selected outbox prefix is empty. */
int32_t potluckdb_runtime_session_prepare_sync_v1(
    PotluckDbRuntimeHandleV1 session_handle,
    uint64_t count,
    uint64_t *out_prepared,
    PotluckDbRuntimeErrorV1 *out_error
);
/* The prepared token remains valid and retryable until explicitly released. */
int32_t potluckdb_runtime_exchange_execute_v1(
    PotluckDbRuntimeHandleV1 exchange_handle,
    uint64_t prepared_token,
    PotluckDbRuntimeSyncOutcomeV1 *out_outcome,
    uint32_t out_outcome_size,
    PotluckDbRuntimeErrorV1 *out_error
);
/* A null upload-options pointer selects a fresh native fixed policy with limit 16. */
int32_t potluckdb_runtime_exchange_execute_with_uploads_v1(
    PotluckDbRuntimeHandleV1 exchange_handle,
    uint64_t prepared_token,
    const PotluckDbRuntimeUploadOptionsV1 *upload_options,
    PotluckDbRuntimeSyncOutcomeV1 *out_outcome,
    uint32_t out_outcome_size,
    PotluckDbRuntimeErrorV1 *out_error
);
/* Apply accepts only a published outcome produced for this session handle. */
int32_t potluckdb_runtime_session_apply_sync_v1(
    PotluckDbRuntimeHandleV1 session_handle,
    uint64_t outcome_token,
    uint32_t *out_applied,
    PotluckDbRuntimeErrorV1 *out_error
);
/* Start performs local preparation and creates one cooperatively driven operation. */
int32_t potluckdb_runtime_session_start_sync_v1(
    PotluckDbRuntimeHandleV1 session_handle,
    uint64_t count,
    PotluckDbRuntimeHandleV1 *out_operation,
    PotluckDbRuntimeErrorV1 *out_error
);
/* Poll performs no storage I/O. Prepared and outcome tokens are operation-owned borrows. */
int32_t potluckdb_runtime_sync_operation_poll_v1(
    PotluckDbRuntimeHandleV1 operation_handle,
    PotluckDbRuntimeSyncOperationStatusV1 *out_status,
    uint32_t out_status_size,
    PotluckDbRuntimeErrorV1 *out_error
);
/* Exchange performs only bucket work and leaves failures retryable. */
int32_t potluckdb_runtime_sync_operation_exchange_v1(
    PotluckDbRuntimeHandleV1 operation_handle,
    PotluckDbRuntimeHandleV1 exchange_handle,
    PotluckDbRuntimeSyncOutcomeV1 *out_outcome,
    uint32_t out_outcome_size,
    PotluckDbRuntimeErrorV1 *out_error
);
int32_t potluckdb_runtime_sync_operation_exchange_with_uploads_v1(
    PotluckDbRuntimeHandleV1 operation_handle,
    PotluckDbRuntimeHandleV1 exchange_handle,
    const PotluckDbRuntimeUploadOptionsV1 *upload_options,
    PotluckDbRuntimeSyncOutcomeV1 *out_outcome,
    uint32_t out_outcome_size,
    PotluckDbRuntimeErrorV1 *out_error
);
/* Complete performs only the matching local apply and is stable after success. */
int32_t potluckdb_runtime_sync_operation_complete_v1(
    PotluckDbRuntimeHandleV1 operation_handle,
    PotluckDbRuntimeSyncOutcomeV1 *out_outcome,
    uint32_t out_outcome_size,
    PotluckDbRuntimeErrorV1 *out_error
);
/* Winner reads run outside the local owner; durable inbox groups run through it. */
int32_t potluckdb_runtime_session_prepare_rebase_v1(
    PotluckDbRuntimeHandleV1 session_handle,
    PotluckDbRuntimeHandleV1 exchange_handle,
    uint64_t prepared_sync_token,
    uint64_t conflict_outcome_token,
    uint64_t *out_prepared_rebase,
    PotluckDbRuntimeErrorV1 *out_error
);
int32_t potluckdb_runtime_session_prepare_resync_v1(
    PotluckDbRuntimeHandleV1 session_handle,
    PotluckDbRuntimeHandleV1 exchange_handle,
    uint64_t *out_prepared_rebase,
    PotluckDbRuntimeErrorV1 *out_error
);
/* A zero token means no completed durable staged inbox was found. */
int32_t potluckdb_runtime_session_recover_prepared_rebase_v1(
    PotluckDbRuntimeHandleV1 session_handle,
    uint64_t *out_prepared_rebase,
    PotluckDbRuntimeErrorV1 *out_error
);
/* Apply performs no bucket operation and accepts only a capability for this session. */
int32_t potluckdb_runtime_session_apply_rebase_v1(
    PotluckDbRuntimeHandleV1 session_handle,
    uint64_t prepared_rebase_token,
    PotluckDbRuntimeRebaseResultV1 *out_result,
    uint32_t out_result_size,
    PotluckDbRuntimeErrorV1 *out_error
);
int32_t potluckdb_runtime_session_clear_quarantines_v1(
    PotluckDbRuntimeHandleV1 session_handle,
    uint64_t *out_count,
    PotluckDbRuntimeErrorV1 *out_error
);
int32_t potluckdb_runtime_exchange_ensure_current_v1(
    PotluckDbRuntimeHandleV1 handle,
    PotluckDbRuntimeErrorV1 *out_error
);
/* Writes an optional current open-head revision; an absent revision is all zeroes. */
int32_t potluckdb_runtime_exchange_current_revision_v1(
    PotluckDbRuntimeHandleV1 handle,
    uint8_t out_revision[32],
    uint32_t *out_has_revision,
    PotluckDbRuntimeErrorV1 *out_error
);
int32_t potluckdb_runtime_exchange_maintain_v1(
    PotluckDbRuntimeHandleV1 handle,
    PotluckDbRuntimeErrorV1 *out_error
);
int32_t potluckdb_runtime_exchange_maintain_report_v1(
    PotluckDbRuntimeHandleV1 handle,
    PotluckDbRuntimeCollectionReportV1 *out_report,
    uint32_t out_report_size,
    PotluckDbRuntimeErrorV1 *out_error
);
int32_t potluckdb_runtime_exchange_collect_v1(
    PotluckDbRuntimeHandleV1 handle,
    uint64_t keep_last,
    uint32_t recover_abandoned,
    PotluckDbRuntimeCollectionReportV1 *out_report,
    uint32_t out_report_size,
    PotluckDbRuntimeErrorV1 *out_error
);
int32_t potluckdb_runtime_exchange_status_v1(
    PotluckDbRuntimeHandleV1 handle,
    PotluckDbRuntimeExchangeStatusV1 *out_status,
    uint32_t out_status_size,
    PotluckDbRuntimeErrorV1 *out_error
);
int32_t potluckdb_runtime_exchange_bucket_counters_v1(
    PotluckDbRuntimeHandleV1 handle,
    PotluckDbRuntimeBucketCounterSnapshotV1 *out_counters,
    uint32_t out_counters_size,
    PotluckDbRuntimeErrorV1 *out_error
);
/* bindings is empty or one SVB1 value; out_rows owns one SVR1 value. */
int32_t potluckdb_runtime_session_execute_v1(
    PotluckDbRuntimeHandleV1 handle,
    PotluckDbRuntimeSliceV1 sql,
    PotluckDbRuntimeSliceV1 bindings,
    PotluckDbRuntimeBufferV1 *out_rows,
    PotluckDbRuntimeErrorV1 *out_error
);
/* Parses exactly one statement without running it. */
int32_t potluckdb_runtime_session_statement_readonly_v1(
    PotluckDbRuntimeHandleV1 handle,
    PotluckDbRuntimeSliceV1 sql,
    uint32_t *out_readonly,
    PotluckDbRuntimeErrorV1 *out_error
);
/* Each call contains exactly one statement and empty or canonical SVB1 bindings. */
/* The managed transaction rolls back if any statement or COMMIT fails. */
int32_t potluckdb_runtime_session_transact_v1(
    PotluckDbRuntimeHandleV1 handle,
    const PotluckDbRuntimeStatementCallV1 *calls,
    uint64_t call_count,
    uint32_t mode,
    PotluckDbRuntimeBufferV1 *out_rows,
    PotluckDbRuntimeErrorV1 *out_error
);
int32_t potluckdb_runtime_session_checkpoint_v1(
    PotluckDbRuntimeHandleV1 handle,
    PotluckDbRuntimeCheckpointResultV1 *out_result,
    uint32_t out_result_size,
    PotluckDbRuntimeErrorV1 *out_error
);
int32_t potluckdb_runtime_session_checkpoint_status_v1(
    PotluckDbRuntimeHandleV1 handle,
    PotluckDbRuntimeCheckpointStatusV1 *out_status,
    uint32_t out_status_size,
    PotluckDbRuntimeErrorV1 *out_error
);
int32_t potluckdb_runtime_session_branch_status_v1(
    PotluckDbRuntimeHandleV1 handle,
    PotluckDbRuntimeBranchStatusV1 *out_status,
    uint32_t out_status_size,
    PotluckDbRuntimeErrorV1 *out_error
);
int32_t potluckdb_runtime_session_prefetch_status_v1(
    PotluckDbRuntimeHandleV1 handle,
    PotluckDbRuntimePrefetchStatusV1 *out_status,
    uint32_t out_status_size,
    PotluckDbRuntimeErrorV1 *out_error
);
int32_t potluckdb_runtime_session_status_v1(
    PotluckDbRuntimeHandleV1 handle,
    PotluckDbRuntimeSessionStatusV1 *out_status,
    uint32_t out_status_size,
    PotluckDbRuntimeErrorV1 *out_error
);
int32_t potluckdb_runtime_connection_execute_v1(
    PotluckDbRuntimeHandleV1 handle,
    PotluckDbRuntimeSliceV1 sql,
    PotluckDbRuntimeSliceV1 bindings,
    PotluckDbRuntimeBufferV1 *out_rows,
    PotluckDbRuntimeErrorV1 *out_error
);
/* Parses exactly one statement without running it. */
int32_t potluckdb_runtime_connection_statement_readonly_v1(
    PotluckDbRuntimeHandleV1 handle,
    PotluckDbRuntimeSliceV1 sql,
    uint32_t *out_readonly,
    PotluckDbRuntimeErrorV1 *out_error
);
int32_t potluckdb_runtime_connection_transact_v1(
    PotluckDbRuntimeHandleV1 handle,
    const PotluckDbRuntimeStatementCallV1 *calls,
    uint64_t call_count,
    uint32_t mode,
    PotluckDbRuntimeBufferV1 *out_rows,
    PotluckDbRuntimeErrorV1 *out_error
);
int32_t potluckdb_runtime_connection_checkpoint_v1(
    PotluckDbRuntimeHandleV1 handle,
    PotluckDbRuntimeCheckpointResultV1 *out_result,
    uint32_t out_result_size,
    PotluckDbRuntimeErrorV1 *out_error
);
int32_t potluckdb_runtime_connection_checkpoint_status_v1(
    PotluckDbRuntimeHandleV1 handle,
    PotluckDbRuntimeCheckpointStatusV1 *out_status,
    uint32_t out_status_size,
    PotluckDbRuntimeErrorV1 *out_error
);
int32_t potluckdb_runtime_connection_prefetch_status_v1(
    PotluckDbRuntimeHandleV1 handle,
    PotluckDbRuntimePrefetchStatusV1 *out_status,
    uint32_t out_status_size,
    PotluckDbRuntimeErrorV1 *out_error
);
int32_t potluckdb_runtime_connection_close_v1(
    PotluckDbRuntimeHandleV1 handle,
    PotluckDbRuntimeErrorV1 *out_error
);
/* close is idempotent; release succeeds only after close and invalidates the handle. */
int32_t potluckdb_runtime_connection_release_v1(
    PotluckDbRuntimeHandleV1 handle,
    PotluckDbRuntimeErrorV1 *out_error
);
int32_t potluckdb_runtime_session_close_v1(
    PotluckDbRuntimeHandleV1 handle,
    PotluckDbRuntimeErrorV1 *out_error
);
/* close is idempotent; release succeeds only after close and invalidates the handle. */
int32_t potluckdb_runtime_session_release_v1(
    PotluckDbRuntimeHandleV1 handle,
    PotluckDbRuntimeErrorV1 *out_error
);
int32_t potluckdb_runtime_exchange_close_v1(
    PotluckDbRuntimeHandleV1 handle,
    PotluckDbRuntimeErrorV1 *out_error
);
/* close is idempotent; release never closes a caller-owned bucket adapter. */
int32_t potluckdb_runtime_exchange_release_v1(
    PotluckDbRuntimeHandleV1 handle,
    PotluckDbRuntimeErrorV1 *out_error
);
int32_t potluckdb_runtime_prepared_release_v1(
    uint64_t token,
    PotluckDbRuntimeErrorV1 *out_error
);
/* Release outcome.reason separately when its buffer token is nonzero. */
int32_t potluckdb_runtime_outcome_release_v1(
    uint64_t token,
    PotluckDbRuntimeErrorV1 *out_error
);
/* Release invalidates the operation and its borrowed prepared/outcome tokens. */
int32_t potluckdb_runtime_sync_operation_release_v1(
    PotluckDbRuntimeHandleV1 operation_handle,
    PotluckDbRuntimeErrorV1 *out_error
);
int32_t potluckdb_runtime_rebase_release_v1(
    uint64_t token,
    PotluckDbRuntimeErrorV1 *out_error
);
int32_t potluckdb_runtime_buffer_release_v1(
    uint64_t token,
    PotluckDbRuntimeErrorV1 *out_error
);

#ifdef __cplusplus
}
#endif

#endif
