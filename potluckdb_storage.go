package potluckdb

/*
#cgo CFLAGS: -I${SRCDIR}
#cgo linux android LDFLAGS: -ldl

#include <stdlib.h>

#include "potluckdb_storage.h"
*/
import "C"

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"unsafe"
)

// S3Options names an S3-compatible bucket and how to reach it.
//
// Built as a struct rather than a pile of arguments so adding a field
// later is not a change to every call site -- the shape this crosses the
// ABI as is JSON for the same reason.
type S3Options struct {
	Bucket          string
	Namespace       string
	Region          string
	Endpoint        string
	ForcePathStyle  bool
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
	Remote          *RemoteOptions
}

// GCSOptions names a Google Cloud Storage bucket and how to reach it.
//
// Built as a struct rather than a pile of arguments so adding a field
// later is not a change to every call site -- the shape this crosses the
// ABI as is JSON for the same reason.
type GCSOptions struct {
	Bucket    string
	Namespace string
	Endpoint  string
	// ServiceAccount is the JSON key Google issues, as its own text.
	// Empty uses Application Default Credentials.
	ServiceAccount string
	Remote         *RemoteOptions
}

// AzureOptions names an Azure Blob Storage container and how to reach
// it.
//
// Built as a struct rather than a pile of arguments so adding a field
// later is not a change to every call site -- the shape this crosses the
// ABI as is JSON for the same reason.
type AzureOptions struct {
	Account   string
	Container string
	Namespace string
	// Endpoint reaches the library as endpoint_url. Azurite is one
	// account under one port, so its endpoint carries the account as a
	// path segment.
	Endpoint string
	// AccountKey is an explicit key. Empty uses the Azure chain.
	AccountKey string
	Remote     *RemoteOptions
}

// storageEntryPoints names every provider the storage library builds,
// and the entry point that builds it.
//
// Closed on purpose: a provider is not a string this SDK forwards to
// dlsym, it is one of three things the library is known to export. A
// name outside this map is refused here, before anything is loaded,
// rather than becoming a symbol lookup that fails in the loader's
// wording instead of this SDK's.
var storageEntryPoints = map[string]string{
	"s3":    "potluckdb_storage_s3_create_v1",
	"gcs":   "potluckdb_storage_gcs_create_v1",
	"azure": "potluckdb_storage_azure_create_v1",
}

func storageUnavailable(message string) *Error {
	return &Error{
		Code:     "storage_library_unavailable",
		Category: CategoryPermanent,
		Message:  message,
		Action: "Ship libpotluckdb_storage beside the program, or set " +
			"POTLUCKDB_STORAGE_LIBRARY to one already on this machine.",
	}
}

// storageUnknownProvider refuses a provider this SDK has no entry point
// for, before the library is looked for: the map above is this SDK's own
// knowledge, and a caller who mistyped a provider should read that back
// rather than a complaint about an absent install.
func storageUnknownProvider(provider string) *Error {
	return &Error{
		Code:     "storage_library_unavailable",
		Category: CategoryPermanent,
		Message: `provider "` + provider + `" is not one this SDK can load. ` +
			`The storage library provides "s3", "gcs", and "azure".`,
		Action: "Correct the provider name and retry.",
	}
}

// storageMissingEntryPoint refuses a provider this SDK knows that the
// loaded library does not export. The ordinary cause is an older storage
// library on the machine than the one this SDK was built against, so the
// remedy is the library, not the call.
func storageMissingEntryPoint(provider string, symbol string) *Error {
	return &Error{
		Code:     "storage_library_unavailable",
		Category: CategoryPermanent,
		Message: `the storage library does not export "` + symbol +
			`". It may predate provider "` + provider +
			`"; update the potluckdb storage library.`,
		Action: "Ship a libpotluckdb_storage matching this SDK's version.",
	}
}

var storageOnce sync.Once
var storageLoadErr error

// storageDefaultLibraryName is the platform's own name for the storage
// library, the one the loader searches for when nothing names a path.
func storageDefaultLibraryName() string {
	switch runtime.GOOS {
	case "windows":
		return "potluckdb_storage.dll"
	case "darwin", "ios":
		return "libpotluckdb_storage.dylib"
	default:
		return "libpotluckdb_storage.so"
	}
}

// storageLibraryPath is where the storage library is for this process:
// the path POTLUCKDB_STORAGE_LIBRARY names, because someone who named a
// library meant that one; then a copy beside the running executable,
// where a Go program's libraries sit, and on macOS one in the Frameworks
// directory of the application bundle the executable is in; otherwise
// the platform's own name, which lets the loader search wherever it
// searches.
func storageLibraryPath() string {
	if named := os.Getenv("POTLUCKDB_STORAGE_LIBRARY"); named != "" {
		return named
	}
	name := storageDefaultLibraryName()
	executable, err := os.Executable()
	if err != nil {
		return name
	}
	if resolved, err := filepath.EvalSymlinks(executable); err == nil {
		executable = resolved
	}
	directory := filepath.Dir(executable)
	candidates := []string{filepath.Join(directory, name)}
	if runtime.GOOS == "darwin" {
		candidates = append(candidates, filepath.Join(directory, "..", "Frameworks", name))
	}
	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	return name
}

func loadStorageLibrary() error {
	storageOnce.Do(func() {
		var buf [1024]byte
		path := C.CString(storageLibraryPath())
		defer C.free(unsafe.Pointer(path))
		if C.potluckdb_go_storage_load(
			path, (*C.char)(unsafe.Pointer(&buf[0])), C.size_t(len(buf)),
		) != 0 {
			message := C.GoString((*C.char)(unsafe.Pointer(&buf[0])))
			if message == "" {
				message = "no potluckdb storage library is available"
			}
			storageLoadErr = storageUnavailable(message)
		}
	})
	return storageLoadErr
}

func (o S3Options) document() (string, error) {
	credentials := map[string]string{"mode": "ambient"}
	if o.AccessKeyID != "" || o.SecretAccessKey != "" {
		credentials = map[string]string{
			"mode":              "explicit",
			"access_key_id":     o.AccessKeyID,
			"secret_access_key": o.SecretAccessKey,
		}
		if o.SessionToken != "" {
			credentials["session_token"] = o.SessionToken
		}
	}
	payload := map[string]any{
		"provider":         "s3",
		"bucket":           o.Bucket,
		"namespace":        o.Namespace,
		"force_path_style": o.ForcePathStyle,
		"credentials":      credentials,
	}
	if o.Region != "" {
		payload["region"] = o.Region
	}
	if o.Endpoint != "" {
		payload["endpoint"] = o.Endpoint
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return "", internalFailure("storage configuration could not be encoded")
	}
	return string(encoded), nil
}

// FromS3 opens a Remote over an S3-compatible bucket, through the shared
// storage library.
//
// Credentials come from the AWS chain -- environment, shared
// configuration and credentials files, SSO, web identity, container and
// instance metadata -- unless AccessKeyID and SecretAccessKey name a
// pair. Pass keys for a store that is not AWS: Cloudflare R2 issues a
// pair with an account Endpoint and wants Region "auto". MinIO and most
// self-hosted gateways want ForcePathStyle, which addresses the bucket
// as a path segment rather than as a subdomain.
func FromS3(options S3Options) (*Remote, error) {
	if options.Bucket == "" || options.Namespace == "" {
		return nil, invalidArgument(
			"an S3 remote needs a bucket and a namespace",
			"Name both: one bucket holds many databases, one per namespace.",
		)
	}
	if (options.AccessKeyID == "") != (options.SecretAccessKey == "") {
		// Half a credential is not a credential and is not an absent one
		// either: taken as absent, the chain would answer with somebody
		// else's keys and the failure would name a permission.
		return nil, invalidArgument(
			"AccessKeyID and SecretAccessKey are given together or not at all",
			"Pass both to use explicit credentials, or neither for the AWS chain.",
		)
	}
	document, err := options.document()
	if err != nil {
		return nil, err
	}
	return fromProvider("s3", document, options.Remote)
}

func (o GCSOptions) document() (string, error) {
	credentials := map[string]any{"mode": "ambient"}
	if o.ServiceAccount != "" {
		// The key crosses as the text it already is: this SDK has a JSON
		// encoder but re-encoding a caller's key would be this facade
		// inventing a shape, and the pack accepts either form.
		credentials = map[string]any{
			"mode":            "explicit",
			"service_account": o.ServiceAccount,
		}
	}
	payload := map[string]any{
		"provider":    "gcs",
		"bucket":      o.Bucket,
		"namespace":   o.Namespace,
		"credentials": credentials,
	}
	if o.Endpoint != "" {
		payload["endpoint"] = o.Endpoint
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return "", internalFailure("storage configuration could not be encoded")
	}
	return string(encoded), nil
}

func (o AzureOptions) document() (string, error) {
	credentials := map[string]any{"mode": "ambient"}
	if o.AccountKey != "" {
		credentials = map[string]any{
			"mode":        "explicit",
			"account_key": o.AccountKey,
		}
	}
	payload := map[string]any{
		"provider":    "azure",
		"account":     o.Account,
		"container":   o.Container,
		"namespace":   o.Namespace,
		"credentials": credentials,
	}
	if o.Endpoint != "" {
		// Endpoint is what every facade calls it and endpoint_url is what
		// the pack reads; the rename happens here so that one spelling
		// reaches callers across all three providers.
		payload["endpoint_url"] = o.Endpoint
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return "", internalFailure("storage configuration could not be encoded")
	}
	return string(encoded), nil
}

// FromGCS opens a Remote over a Google Cloud Storage bucket, through the
// shared storage library.
//
// Credentials come from Application Default Credentials --
// GOOGLE_APPLICATION_CREDENTIALS, the gcloud user credentials file, then
// the metadata server -- unless ServiceAccount names a key. Pass the
// JSON key Google issues for a process that has no such chain to
// consult.
func FromGCS(options GCSOptions) (*Remote, error) {
	if options.Bucket == "" || options.Namespace == "" {
		return nil, invalidArgument(
			"a GCS remote needs a bucket and a namespace",
			"Name both: one bucket holds many databases, one per namespace.",
		)
	}
	document, err := options.document()
	if err != nil {
		return nil, err
	}
	return fromProvider("gcs", document, options.Remote)
}

// FromAzure opens a Remote over an Azure Blob Storage container, through
// the shared storage library.
//
// Credentials come from the Azure chain -- environment, workload
// identity, managed identity, the Azure CLI and the developer tools --
// unless AccountKey names a key. Azurite issues the well-known
// devstoreaccount1 pair and wants an Endpoint that carries the account.
func FromAzure(options AzureOptions) (*Remote, error) {
	if options.Account == "" || options.Container == "" || options.Namespace == "" {
		return nil, invalidArgument(
			"an Azure remote needs an account, a container and a namespace",
			"Name all three: one container holds many databases, one per namespace.",
		)
	}
	document, err := options.document()
	if err != nil {
		return nil, err
	}
	return fromProvider("azure", document, options.Remote)
}

// fromProvider opens a Remote over whichever provider's adapter the
// storage library builds.
//
// The one place this SDK crosses into the storage library. Everything
// that is not the provider -- the load, the ownership of the adapter
// across the exchange, the refusal the library answers a bad
// configuration with -- is the same for all three, so it is written once
// here and each facade supplies only its own document.
//
// Unexported: naming a provider and hand-building its document is not
// the surface callers get. FromS3, FromGCS and FromAzure are.
func fromProvider(
	provider string, document string, options *RemoteOptions,
) (*Remote, error) {
	symbol, known := storageEntryPoints[provider]
	if !known {
		// Before loadStorageLibrary, deliberately: this is answerable
		// without a library and the answer is more useful than
		// "none installed".
		return nil, storageUnknownProvider(provider)
	}
	selected, err := options.normalized()
	if err != nil {
		return nil, err
	}
	if err := loadStorageLibrary(); err != nil {
		return nil, err
	}

	entryPoint := C.CString(symbol)
	defer C.free(unsafe.Pointer(entryPoint))
	payload := C.CBytes([]byte(document))
	defer C.free(payload)
	var adapter *C.PotluckDbRuntimeBucketAdapterV1
	var reported [2048]byte
	code := C.potluckdb_go_storage_create(
		entryPoint,
		(*C.uint8_t)(payload),
		C.uint64_t(len(document)),
		&adapter,
		(*C.char)(unsafe.Pointer(&reported[0])),
		C.size_t(len(reported)),
	)
	runtime.KeepAlive(reported)
	if code == C.POTLUCKDB_GO_STORAGE_NO_ENTRY_POINT_V1 {
		return nil, storageMissingEntryPoint(provider, symbol)
	}
	if code != C.POTLUCKDB_STORAGE_OK_V1 {
		message := C.GoString((*C.char)(unsafe.Pointer(&reported[0])))
		if message == "" {
			message = "the storage library refused the configuration"
		}
		if code == 24 {
			return nil, &Error{Code: "bucket_not_found", Category: CategoryPermanent,
				Message: message, Action: "Create the bucket or container, or correct the name the remote was configured with."}
		}
		return nil, invalidArgument(
			message,
			"Correct the storage configuration and retry.",
		)
	}

	storage, err := RetainStorageAdapter(unsafe.Pointer(adapter), nil)
	if err != nil {
		C.potluckdb_go_storage_release(adapter)
		return nil, err
	}
	remote, err := RemoteFromStorage(storage, selected)
	if err != nil {
		// An exchange that did not open is not holding the adapter, so
		// the adapter goes back rather than leaking a thread pool.
		C.potluckdb_go_storage_release(adapter)
		return nil, err
	}
	held := adapter
	remote.state.afterClose = func() {
		C.potluckdb_go_storage_release(held)
	}
	return remote, nil
}
