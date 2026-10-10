// Package s3 is an S3 bucket adapter: the seam's semantics, over a
// client you supply.
//
// The runtime deliberately holds no storage client -- bucket I/O arrives
// through host callbacks so each SDK brings its own HTTP stack,
// credentials and platform integration. This package keeps that decision
// and pays the other half of it.
//
// What it does not do is speak HTTP. [S3Client] is an interface, and an
// application hands in whatever it already uses -- the AWS SDK for Go, a
// hand-rolled signer, a test double. That mirrors adapters/s3.py and
// native/sdk/rust/src/s3.rs, which take the same shape for the same
// reason: an SDK that shipped its own HTTP and SigV4 stack would be
// making a size and dependency decision on behalf of every application
// that links it. Nothing here imports anything outside the standard
// library.
//
// What it does do is the part that is hard to get right and easy to get
// wrong quietly (#67). Conditional writes, compare-and-swap over whole
// values, idempotent puts that tell agreement from disagreement, lexical
// listing, and the error mapping the seam requires. An application that
// writes this itself can pass every test it thinks to run and still lose
// revisions under contention.
//
// The split with the client is deliberate: a client performs one request
// and reports its outcome, and everything spanning requests -- retrying a
// conditional write, walking a listing to its last page, deciding what a
// 412 means -- lives here, where it is written once rather than once per
// application. contracts/seam-4-bucket.md is what it implements; B7, B8,
// B12, B13 and B15 are the requirements that shape this file.
//
// # Reaching the runtime
//
// This package is separate from the potluckdb facade, and it does not
// import it: the facade is cgo over the native library, so importing it
// would make an adapter that needs no runtime require one to compile.
// Remote.FromStorage on this SDK takes a pointer to a native
// PotluckDbRuntimeBucketAdapterV1 whose callbacks a caller fills in.
// Nothing in this SDK yet builds that table from a Go implementor, so
// this adapter is not something Remote.FromStorage can be handed
// directly -- see the package README note in the pull request that added
// it. The semantics below are what such a bridge would carry.
package s3

import (
	"container/list"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// ConditionalRetryLimit is how many times a conditional write may lose
// its race before the adapter stops rather than looping.
//
// A swap that keeps losing is a busy cell, which is ordinary; a swap that
// loses this many times in a row is a store that is not making progress,
// and saying so beats spinning.
const ConditionalRetryLimit = 8

// ImmutableObservationLimit is how many immutable observations may be
// retained at once, which is the ceiling B8 sets.
//
// The cache is what lets a repeated put of bytes this adapter already
// wrote skip the body. Unbounded it would grow with every distinct key a
// long-lived process touches, so the oldest observation goes when a new
// one arrives at the limit.
const ImmutableObservationLimit = 1024

// keyByteLimit is the longest key the seam accepts, in bytes.
const keyByteLimit = 160

// probeWrongETag is an entity tag no object carries: the one the B15
// probe's swap names, which a store enforcing If-Match refuses.
const probeWrongETag = `"00000000000000000000000000000000"`

// The three B15 probe bodies: 33, 48 and 45 bytes, 126 in total. Stated
// as strings and converted per probe, so no caller can mutate a shared
// body through the client it was handed to.
const (
	probeCreatedBody = "potluckdb conditional-write probe"
	probeSecondBody  = "potluckdb conditional-write probe: second create"
	probeStaleBody   = "potluckdb conditional-write probe: stale swap"
)

// The Seam-4 error codes, which are the runtime's own
// (POTLUCKDB_RUNTIME_ERROR_* in native/interface/include/potluckdb_runtime.h).
// A bucket adapter is called by the runtime across Seam 4 rather than by
// an application, so it answers in the runtime's vocabulary and not in
// the string codes the potluckdb facade reports to callers.
const (
	ErrorInternalFailure    int32 = 1
	ErrorMissing            int32 = 2
	ErrorConflict           int32 = 6
	ErrorUnavailable        int32 = 7
	ErrorContractViolation  int32 = 18
	ErrorStorageUnsupported int32 = 23
	ErrorBucketNotFound     int32 = 24
)

// Error is a typed Seam-4 failure.
type Error struct {
	Code    int32
	Message string
}

func (e *Error) Error() string { return e.Message }

// The outcomes this adapter reasons about, in the terms an [S3Client]
// reports them.
//
// A client fails in its own vocabulary -- an SDK error type, an HTTP
// status, a code string. Translating once, at the client boundary, keeps
// the semantics below free of any one client's shape. Wrap these rather
// than returning them bare when there is detail worth keeping:
// fmt.Errorf("get %q: %w", key, s3.ErrMissing) classifies correctly and
// still says what happened.
//
// Any other non-nil error is treated as [ErrUnavailable]. That default is
// the safe one: an unrecognized failure must never read as a satisfied
// condition.
var (
	// ErrMissing is no such key: 404, NoSuchKey, NotFound.
	ErrMissing      = errors.New("s3: no such key")
	ErrNoSuchBucket = errors.New("s3: no such bucket")
	// ErrPrecondition is a conditional header that was not satisfied:
	// 412, or the 409 that S3 returns when two If-None-Match: * writes
	// race.
	ErrPrecondition = errors.New("s3: conditional request failed")
	// ErrUnavailable is anything else: throttling, network, permissions,
	// a bad request.
	ErrUnavailable = errors.New("s3: object store is unavailable")
	// ErrUnsupported is a conditional header the store does not
	// implement: 501 Not Implemented. No retry changes that answer, so
	// the write is refused as unsupported storage (B15) rather than
	// unavailable.
	ErrUnsupported = errors.New("s3: conditional writes are not implemented")
)

// S3Object is one object, as the store returned it.
type S3Object struct {
	Data []byte
	// ETag is the entity tag, used to make the next conditional write
	// refer to exactly the value that was read. A read that arrives
	// without one is refused as Unavailable: there would be nothing to
	// name in the If-Match that follows.
	ETag string
}

// S3Head is what HEAD reported about an object.
//
// Both fields are needed together: B8 lets a remembered observation stand
// in for a body only when the tag *and* the length still match, so a
// client that reports one without the other cannot support the
// idempotent-put shortcut.
type S3Head struct {
	ETag string
	// Size is Content-Length, in bytes.
	Size int64
}

// S3Page is one page of a listing, as the store returned it.
type S3Page struct {
	// Keys are the keys on this page, in any order. The adapter sorts and
	// de-duplicates across every page.
	Keys []string
	// NextToken asks for the next page. Empty means this page is the last
	// one.
	NextToken string
	// Truncated is the store's own IsTruncated. It exists to catch the
	// malformed answer a bare token cannot express -- more pages promised
	// and none named -- which would otherwise truncate a listing in
	// silence. Pagination follows NextToken, so a client that fills in
	// the token and forgets this field still walks every page.
	Truncated bool
}

// S3Client is the transport this adapter drives.
//
// Every method performs exactly one request and returns the store's own
// outcome, classified against [ErrMissing], [ErrPrecondition],
// [ErrUnavailable] and [ErrUnsupported]. Implementations do not need to
// know the seam: they
// need to put bytes somewhere, honour two conditional headers, and hand
// back one page at a time.
//
// B10, B11 and B13 have the adapter issue reads, writes and the cell pair
// concurrently on one instance, so an implementation must be safe for
// concurrent use.
type S3Client interface {
	GetObject(key string) (S3Object, error)

	// PutObjectIfAbsent is PUT with If-None-Match: * -- it succeeds only
	// if the key is absent, and returns the new entity tag.
	//
	// This is what makes an immutable put safe. Without it two writers
	// racing on the same key both succeed and the later one wins, which
	// for immutable objects means a disagreement nobody notices.
	PutObjectIfAbsent(key string, data []byte) (string, error)

	// PutObjectIfMatch is PUT with If-Match: <etag> -- it succeeds only if
	// the current value is exactly the one that etag came from, and
	// returns the new entity tag.
	//
	// This is the compare-and-swap. A store that accepts this header and
	// ignores it (which is what Google Cloud Storage does through its
	// S3-compatible endpoint, see #57) makes every swap succeed and loses
	// revisions silently. Do not point this adapter at one.
	PutObjectIfMatch(key string, data []byte, etag string) (string, error)

	// HeadObject is HEAD, returning the current entity tag and byte
	// length.
	//
	// The pair is what makes the tag worth having: it is how a remembered
	// observation is checked for currency without fetching the bytes.
	HeadObject(key string) (S3Head, error)

	// ListObjects is one ListObjectsV2 request: the keys under prefix
	// that this page carries, and the token that continues the listing.
	//
	// continuation is empty for the first request and afterwards the
	// NextToken of the page before it. Walking the pages is the adapter's
	// job, not the client's -- a listing that stops at the provider's page
	// limit is a listing that silently omits objects, and no implementor
	// of this interface should have to know that.
	ListObjects(prefix, continuation string) (S3Page, error)

	DeleteObject(key string) error
}

// cellObservation is what a cell looked like when this adapter last saw
// it.
//
// Kept so a swap can name the exact value it intends to replace. It is a
// cache of an observation, never a source of truth: every path that uses
// one is prepared for it to be stale, and re-reads. There are two of
// these at most, so the whole value is affordable -- which is what B8
// asks for, since the swap compares the caller's expected bytes against
// it.
type cellObservation struct {
	value []byte
	etag  string
}

// immutableObservation is what an immutable object looked like when this
// adapter last saw it.
//
// B8 permits key, entity tag, byte length and SHA-256, and no more. The
// bytes themselves are deliberately absent: an adapter that kept them
// would hold every object it has written for as long as it lives, and the
// fingerprint answers the only question asked of it -- whether a candidate
// is the same value.
type immutableObservation struct {
	etag        string
	size        int64
	fingerprint [32]byte
}

// cacheEntry carries its own key so eviction can find the map entry from
// the recency list, and keeps recency beside the observation rather than
// inside it, so an observation stays the four fields B8 allows.
type cacheEntry struct {
	key         string
	observation immutableObservation
}

// observationCache holds the retained observations, bounded by
// [ImmutableObservationLimit], with the least recently used at the front.
type observationCache struct {
	entries map[string]*list.Element
	recency *list.List
}

func newObservationCache() *observationCache {
	return &observationCache{
		entries: make(map[string]*list.Element),
		recency: list.New(),
	}
}

func (c *observationCache) remember(key string, observation immutableObservation) {
	if element, found := c.entries[key]; found {
		element.Value.(*cacheEntry).observation = observation
		c.recency.MoveToBack(element)
		return
	}
	c.entries[key] = c.recency.PushBack(&cacheEntry{key: key, observation: observation})
	for len(c.entries) > ImmutableObservationLimit {
		oldest := c.recency.Front()
		if oldest == nil {
			break
		}
		c.recency.Remove(oldest)
		delete(c.entries, oldest.Value.(*cacheEntry).key)
	}
}

// matching returns the observation for key, if it describes exactly these
// bytes.
//
// Length and fingerprint together: the length alone would accept a
// different value of the same size, which is the disagreement B1 exists
// to catch.
func (c *observationCache) matching(
	key string, size int64, fingerprint *[32]byte,
) (immutableObservation, bool) {
	element, found := c.entries[key]
	if !found {
		return immutableObservation{}, false
	}
	observation := element.Value.(*cacheEntry).observation
	if observation.size != size || !sameBytes(observation.fingerprint[:], fingerprint[:]) {
		return immutableObservation{}, false
	}
	c.recency.MoveToBack(element)
	return observation, true
}

func (c *observationCache) forget(key string) {
	if element, found := c.entries[key]; found {
		c.recency.Remove(element)
		delete(c.entries, key)
	}
}

func (c *observationCache) clear() {
	c.entries = make(map[string]*list.Element)
	c.recency.Init()
}

type counters struct {
	gets                    uint64
	puts                    uint64
	idempotentPuts          uint64
	exists                  uint64
	lists                   uint64
	deletes                 uint64
	cellGets                uint64
	cellSwaps               uint64
	cellSwapWins            uint64
	transportGets           uint64
	transportPuts           uint64
	transportHeads          uint64
	transportLists          uint64
	transportDeletes        uint64
	transportBytesReceived  uint64
	transportBytesSubmitted uint64
}

// BucketCounterSnapshot is the observational adapter-counter snapshot
// Seam 4 defines. B12 separates the logical Seam-4 operations from the
// provider API calls they cause.
type BucketCounterSnapshot struct {
	Gets                    uint64
	Puts                    uint64
	IdempotentPuts          uint64
	Exists                  uint64
	Lists                   uint64
	Deletes                 uint64
	CellGets                uint64
	CellSwaps               uint64
	CellSwapWins            uint64
	TransportGets           uint64
	TransportPuts           uint64
	TransportHeads          uint64
	TransportLists          uint64
	TransportDeletes        uint64
	TransportBytesReceived  uint64
	TransportBytesSubmitted uint64
}

// S3BucketStorage is a bucket over S3-compatible object storage.
//
// The zero value is not usable; construct one with [New]. It is safe for
// concurrent use, as B10, B11 and B13 require.
type S3BucketStorage struct {
	client    S3Client
	namespace string

	// mu guards everything below it. It is never held across a client
	// call: holding it across one would serialize the concurrent gets and
	// puts B10 and B11 require, and the pair reads of B13.
	mu       sync.Mutex
	closed   bool
	counters counters
	// conditionsVerified is whether the B15 probe has found the store
	// refusing both conditional writes. Only a pass is remembered.
	conditionsVerified bool
	// cells holds cell observations by cell name. Two entries at most.
	cells map[string]cellObservation
	// objects holds immutable-object observations, so a repeated put of
	// bytes this adapter already wrote need not send them again.
	objects *observationCache
}

func missing(what string) error {
	return &Error{ErrorMissing, fmt.Sprintf("bucket object is absent: %s", what)}
}

func conflict(what string) error {
	return &Error{
		ErrorConflict,
		fmt.Sprintf("immutable bucket object disagrees: %s", what),
	}
}

func contract(message string) error {
	return &Error{ErrorContractViolation, message}
}

func unavailable(operation string, causes ...error) error {
	for _, cause := range causes {
		if errors.Is(cause, ErrNoSuchBucket) {
			return &Error{ErrorBucketNotFound, "S3 bucket does not exist (NoSuchBucket)"}
		}
	}
	return &Error{
		ErrorUnavailable,
		fmt.Sprintf("S3 bucket storage failed during %s", operation),
	}
}

// unavailableResponse is an Unavailable that names what the store did
// rather than which operation was running, for the answers no retry can
// fix.
func unavailableResponse(detail string) error {
	return &Error{ErrorUnavailable, fmt.Sprintf("S3 bucket storage: %s", detail)}
}

// unsupported is B15: the store accepted, or does not implement, a
// conditional write that every swap depends on it refusing.
func unsupported(detail string) error {
	return &Error{
		ErrorStorageUnsupported,
		fmt.Sprintf(
			"the object store does not enforce conditional writes: %s; "+
				"potluckdb creates no database on it",
			detail,
		),
	}
}

// probeSequence keeps probe names distinct within a process.
var probeSequence atomic.Uint64

// probeName is a name no other probe uses, so probes from devices creating
// the same remote at once cannot answer for each other. Unique, not
// secret.
func probeName() string {
	var scratch [8]byte
	digest := sha256.New()
	binary.LittleEndian.PutUint64(scratch[:], uint64(os.Getpid()))
	digest.Write(scratch[:])
	binary.LittleEndian.PutUint64(scratch[:], probeSequence.Add(1))
	digest.Write(scratch[:])
	binary.LittleEndian.PutUint64(scratch[:], uint64(time.Now().UnixNano()))
	digest.Write(scratch[:])
	return hex.EncodeToString(digest.Sum(nil)[:16])
}

// sameBytes is a constant-time comparison, so a value cannot be recovered
// by timing how far a comparison got. These are content-addressed bytes
// and the risk is remote, but the cost is a loop that does not stop
// early. ConstantTimeCompare already answers 0 for unequal lengths.
func sameBytes(left, right []byte) bool {
	return subtle.ConstantTimeCompare(left, right) == 1
}

// fingerprintOf is the length and SHA-256 that stand in for a value in
// the cache.
func fingerprintOf(data []byte) (int64, [32]byte) {
	return int64(len(data)), sha256.Sum256(data)
}

// New returns a bucket over client, with every key under namespace.
//
// The namespace is a prefix, not a bucket: one bucket can hold many
// independent databases, and nothing here reads outside its own. The
// client stays the caller's -- [S3BucketStorage.Close] does not close it.
func New(client S3Client, namespace string) (*S3BucketStorage, error) {
	if client == nil {
		return nil, contract("bucket client must not be nil")
	}
	// One trailing separator is how a prefix is usually written, and this
	// package supplies its own; anything else has to be a valid name, so
	// a namespace cannot reach out of itself either.
	namespace = strings.TrimRight(namespace, "/")
	if namespace == "" {
		return nil, contract("bucket namespace must not be empty")
	}
	if err := validateName("namespace", namespace); err != nil {
		return nil, err
	}
	return &S3BucketStorage{
		client:    client,
		namespace: namespace,
		cells:     make(map[string]cellObservation),
		objects:   newObservationCache(),
	}, nil
}

// Close closes the adapter, idempotently (B7).
//
// Later operations answer Unavailable. The client is the caller's: B8
// requires that closing the adapter leave it open, and nothing here
// touches it. Callers quiesce their own concurrent operations first, as
// B10 and B11 require.
//
// The error is always nil; Close reports one so that an S3BucketStorage
// satisfies io.Closer.
func (s *S3BucketStorage) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.closed = true
	s.cells = make(map[string]cellObservation)
	s.objects.clear()
	return nil
}

func (s *S3BucketStorage) objectRoot() string {
	return s.namespace + "/objects/"
}

func (s *S3BucketStorage) objectKey(key string) string {
	return s.objectRoot() + key
}

func (s *S3BucketStorage) cellKey(name string) string {
	return s.namespace + "/cells/" + name
}

// withState runs body under the state lock, refusing every operation once
// the adapter is closed.
func (s *S3BucketStorage) withState(body func()) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return unavailableResponse("the adapter is closed")
	}
	body()
	return nil
}

// count records transport or logical work.
//
// The lock is taken around the counter and released again before any
// client call, for the reason given on the mutex.
func (s *S3BucketStorage) count(update func(*counters)) error {
	return s.withState(func() { update(&s.counters) })
}

// validateName applies the key rules potluckdb_runtime.bucket.validate_key
// applies, which every Seam-4 adapter shares: a portable relative ASCII
// name of at most 160 bytes, opening with a letter or digit, built from
// letters, digits, '.', '_', '-' and '/', with no empty, "." or ".."
// segment.
//
// B7 decides this from the name alone. Nothing here asks the store what
// exists.
func validateName(what, name string) error {
	if name == "" || len(name) > keyByteLimit {
		return contract(fmt.Sprintf(
			"bucket %s must be 1 to %d bytes", what, keyByteLimit,
		))
	}
	if !isASCIIAlphanumeric(name[0]) {
		return contract(fmt.Sprintf(
			"bucket %s must begin with an ASCII letter or digit", what,
		))
	}
	for index := 0; index < len(name); index++ {
		character := name[index]
		if isASCIIAlphanumeric(character) {
			continue
		}
		if character != '.' && character != '_' && character != '-' && character != '/' {
			return contract(fmt.Sprintf(
				"bucket %s must be a portable relative ASCII name", what,
			))
		}
	}
	// An empty segment is a doubled or trailing separator; "." and ".."
	// are traversal. Either would let a valid-looking key name something
	// outside the objects it is allowed to reach.
	for _, segment := range strings.Split(name, "/") {
		if segment == "" || segment == "." || segment == ".." {
			return contract(fmt.Sprintf(
				"bucket %s contains an empty, . or .. segment", what,
			))
		}
	}
	return nil
}

func isASCIIAlphanumeric(character byte) bool {
	return (character >= '0' && character <= '9') ||
		(character >= 'a' && character <= 'z') ||
		(character >= 'A' && character <= 'Z')
}

func validateKey(key string) error { return validateName("key", key) }

// validatePrefix accepts a key, a key with one trailing separator naming
// a subtree, or the empty prefix meaning everything.
func validatePrefix(prefix string) error {
	if prefix == "" {
		return nil
	}
	return validateName("prefix", strings.TrimSuffix(prefix, "/"))
}

// isMissing, isPrecondition and isUnsupported classify a client's
// failure. Everything they all refuse is unavailable: a failure that
// cannot be read must never pass as a satisfied condition.
func isMissing(err error) bool      { return errors.Is(err, ErrMissing) }
func isPrecondition(err error) bool { return errors.Is(err, ErrPrecondition) }
func isUnsupported(err error) bool  { return errors.Is(err, ErrUnsupported) }

// read reads an object, recording what was transferred.
func (s *S3BucketStorage) read(physical, name string) (S3Object, error) {
	if err := s.count(func(c *counters) { c.transportGets++ }); err != nil {
		return S3Object{}, err
	}
	object, err := s.client.GetObject(physical)
	if err != nil {
		if isMissing(err) {
			return S3Object{}, missing(name)
		}
		return S3Object{}, unavailable("object read", err)
	}
	if object.ETag == "" {
		return S3Object{}, unavailableResponse("a read answered without an entity tag")
	}
	if err := s.count(func(c *counters) {
		c.transportBytesReceived += uint64(len(object.Data))
	}); err != nil {
		return S3Object{}, err
	}
	return object, nil
}

// forgetCell drops a remembered cell whether or not the adapter is still
// open, because a closed adapter has nothing worth keeping either.
func (s *S3BucketStorage) forgetCell(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.cells, name)
}

func (s *S3BucketStorage) cellGet(name string) ([]byte, error) {
	object, err := s.read(s.cellKey(name), name)
	if err != nil {
		// A cell that is gone invalidates what we remembered of it: a
		// later swap must not name an etag for a value that no longer
		// exists.
		s.forgetCell(name)
		return nil, err
	}
	// The observation keeps its own copy: the caller owns what it is
	// handed, and B1 does not allow adapter state to alias it.
	remembered := append([]byte(nil), object.Data...)
	if err := s.withState(func() {
		s.counters.cellGets++
		s.cells[name] = cellObservation{value: remembered, etag: object.ETag}
	}); err != nil {
		return nil, err
	}
	return object.Data, nil
}

func (s *S3BucketStorage) optionalCellGet(name string) ([]byte, error) {
	value, err := s.cellGet(name)
	if err != nil {
		var typed *Error
		if errors.As(err, &typed) && typed.Code == ErrorMissing {
			return nil, nil
		}
		return nil, err
	}
	return value, nil
}

// observeCell reads the cell and remembers it, or forgets it if it is
// absent.
func (s *S3BucketStorage) observeCell(name string) (*cellObservation, error) {
	object, err := s.read(s.cellKey(name), name)
	if err != nil {
		var typed *Error
		if errors.As(err, &typed) && typed.Code == ErrorMissing {
			if err := s.withState(func() { delete(s.cells, name) }); err != nil {
				return nil, err
			}
			return nil, nil
		}
		return nil, err
	}
	observation := cellObservation{
		value: append([]byte(nil), object.Data...),
		etag:  object.ETag,
	}
	if err := s.withState(func() { s.cells[name] = observation }); err != nil {
		return nil, err
	}
	return &observation, nil
}

// recordCellSwapWin counts a won swap and remembers what it installed --
// or forgets the cell when the store answered without a tag, since there
// would be nothing to name in the next If-Match.
func (s *S3BucketStorage) recordCellSwapWin(name string, value []byte, etag string) error {
	return s.withState(func() {
		if etag == "" {
			delete(s.cells, name)
		} else {
			s.cells[name] = cellObservation{
				value: append([]byte(nil), value...),
				etag:  etag,
			}
		}
		s.counters.cellSwapWins++
	})
}

// verifyConditionalWrites is B15: before this adapter creates a cell,
// the store refuses the two conditional writes every swap depends on.
//
// Creating a cell is creating a database -- genesis, or a rotation into a
// fresh target -- so this is the last moment to find a store that ignores
// the conditions, before two devices can each install a head and each
// believe they won. The probe writes one object under probe/, outside the
// objects and cells, and deletes it.
func (s *S3BucketStorage) verifyConditionalWrites() error {
	var verified bool
	if err := s.withState(func() { verified = s.conditionsVerified }); err != nil {
		return err
	}
	if verified {
		return nil
	}
	probe := s.namespace + "/probe/" + probeName()
	created := []byte(probeCreatedBody)
	if err := s.countProbePut(len(created)); err != nil {
		return err
	}
	if _, err := s.client.PutObjectIfAbsent(probe, created); err != nil {
		if isUnsupported(err) {
			return unsupported("it answered If-None-Match: * with 501 Not Implemented")
		}
		return unavailable("conditional-write probe", err)
	}
	verdict := s.probeRefusals(probe)
	// Best effort: a probe object left behind by a failed delete is a few
	// bytes outside every name the database reads, and the verdict stands
	// either way.
	if err := s.count(func(c *counters) { c.transportDeletes++ }); err != nil {
		return err
	}
	_ = s.client.DeleteObject(probe)
	if verdict != nil {
		return verdict
	}
	return s.withState(func() { s.conditionsVerified = true })
}

// probeRefusals are the two refusals, against the probe object that now
// exists.
func (s *S3BucketStorage) probeRefusals(probe string) error {
	second := []byte(probeSecondBody)
	if err := s.countProbePut(len(second)); err != nil {
		return err
	}
	if _, err := s.client.PutObjectIfAbsent(probe, second); err == nil {
		return unsupported(
			"it accepted a second create of an existing object under If-None-Match: *",
		)
	} else if !isMissing(err) && !isPrecondition(err) {
		if isUnsupported(err) {
			return unsupported("it answered If-None-Match: * with 501 Not Implemented")
		}
		return unavailable("conditional-write probe", err)
	}
	stale := []byte(probeStaleBody)
	if err := s.countProbePut(len(stale)); err != nil {
		return err
	}
	if _, err := s.client.PutObjectIfMatch(probe, stale, probeWrongETag); err == nil {
		return unsupported(
			"it accepted a replacement under an If-Match naming another entity tag",
		)
	} else if !isMissing(err) && !isPrecondition(err) {
		if isUnsupported(err) {
			return unsupported("it answered If-Match with 501 Not Implemented")
		}
		return unavailable("conditional-write probe", err)
	}
	return nil
}

func (s *S3BucketStorage) countProbePut(size int) error {
	return s.count(func(c *counters) {
		c.transportPuts++
		c.transportBytesSubmitted += uint64(size)
	})
}

func (s *S3BucketStorage) cellSwap(name string, expected, replacement []byte) (bool, error) {
	physical := s.cellKey(name)
	if err := s.count(func(c *counters) { c.cellSwaps++ }); err != nil {
		return false, err
	}

	// Installing where nothing is: the conditional header is "absent",
	// and losing means somebody else got there first.
	if expected == nil {
		if err := s.verifyConditionalWrites(); err != nil {
			return false, err
		}
		if err := s.count(func(c *counters) {
			c.transportPuts++
			c.transportBytesSubmitted += uint64(len(replacement))
		}); err != nil {
			return false, err
		}
		etag, err := s.client.PutObjectIfAbsent(physical, replacement)
		switch {
		case err == nil:
			if err := s.recordCellSwapWin(name, replacement, etag); err != nil {
				return false, err
			}
			return true, nil
		case isMissing(err) || isPrecondition(err):
			if err := s.withState(func() { delete(s.cells, name) }); err != nil {
				return false, err
			}
			return false, nil
		case isUnsupported(err):
			return false, unsupported(
				"it answered If-None-Match: * with 501 Not Implemented",
			)
		default:
			return false, unavailable("cell swap", err)
		}
	}

	// Replacing a specific value. The etag has to come from a read that
	// saw exactly `expected`, or the swap would be conditional on
	// something the caller did not ask about.
	var observation *cellObservation
	s.mu.Lock()
	if seen, found := s.cells[name]; found {
		copied := seen
		observation = &copied
	}
	s.mu.Unlock()
	if observation == nil || !sameBytes(observation.value, expected) {
		fresh, err := s.observeCell(name)
		if err != nil {
			return false, err
		}
		if fresh == nil || !sameBytes(fresh.value, expected) {
			return false, nil
		}
		observation = fresh
	}

	for attempt := 0; attempt < ConditionalRetryLimit; attempt++ {
		if err := s.count(func(c *counters) {
			c.transportPuts++
			c.transportBytesSubmitted += uint64(len(replacement))
		}); err != nil {
			return false, err
		}
		etag, err := s.client.PutObjectIfMatch(physical, replacement, observation.etag)
		if err == nil {
			if err := s.recordCellSwapWin(name, replacement, etag); err != nil {
				return false, err
			}
			return true, nil
		}
		if isUnsupported(err) {
			return false, unsupported("it answered If-Match with 501 Not Implemented")
		}
		if !isMissing(err) && !isPrecondition(err) {
			return false, unavailable("cell swap", err)
		}
		// The value moved under us. It may have moved back: a cell can
		// return to earlier bytes, and then this swap is still the one
		// the caller asked for. Re-read and decide again rather than
		// assuming a loss.
		fresh, observeErr := s.observeCell(name)
		if observeErr != nil {
			return false, observeErr
		}
		if fresh == nil || !sameBytes(fresh.value, expected) {
			return false, nil
		}
		observation = fresh
	}
	// Out of attempts: the cell keeps returning to the expected bytes and
	// moving away again before this write lands. That is a lost
	// compare-and-swap, not a broken store, and the caller retries from a
	// fresh observation.
	if err := s.withState(func() { delete(s.cells, name) }); err != nil {
		return false, err
	}
	return false, nil
}

// immutableObservationIsCurrent reports whether the store still holds
// exactly the object this observation describes.
//
// B8 requires both halves: an entity tag alone would accept a value some
// stores tag by content, and a length alone accepts any value of the same
// size.
func (s *S3BucketStorage) immutableObservationIsCurrent(
	physical, key string, seen immutableObservation,
) (bool, error) {
	if err := s.count(func(c *counters) { c.transportHeads++ }); err != nil {
		return false, err
	}
	current, err := s.client.HeadObject(physical)
	if err != nil {
		if isMissing(err) {
			if err := s.withState(func() { s.objects.forget(key) }); err != nil {
				return false, err
			}
			return false, nil
		}
		return false, unavailable("head immutable", err)
	}
	if current.Size == seen.size && sameBytes([]byte(current.ETag), []byte(seen.etag)) {
		return true, nil
	}
	// Gone, or replaced by something else. Either way the memory is
	// worthless: forget it and write.
	if err := s.withState(func() { s.objects.forget(key) }); err != nil {
		return false, err
	}
	return false, nil
}

// rememberImmutable remembers an object by the fingerprint the caller
// already took, so a put hashes its payload once.
func (s *S3BucketStorage) rememberImmutable(
	key string, size int64, fingerprint [32]byte, etag string,
) error {
	// A store that answers a write without a tag leaves nothing to
	// revalidate against, so there is nothing worth remembering.
	if etag == "" {
		return nil
	}
	return s.withState(func() {
		s.objects.remember(key, immutableObservation{
			etag:        etag,
			size:        size,
			fingerprint: fingerprint,
		})
	})
}

// Get reads an immutable object, reporting Missing when it is absent.
func (s *S3BucketStorage) Get(key string) ([]byte, error) {
	if err := validateKey(key); err != nil {
		return nil, err
	}
	object, err := s.read(s.objectKey(key), key)
	if err != nil {
		return nil, err
	}
	if err := s.count(func(c *counters) { c.gets++ }); err != nil {
		return nil, err
	}
	return object.Data, nil
}

// Put creates an immutable object, or idempotently verifies one that is
// already there. Different bytes under a key that exists are a Conflict,
// and the first value survives (B1).
func (s *S3BucketStorage) Put(key string, data []byte) error {
	if err := validateKey(key); err != nil {
		return err
	}
	physical := s.objectKey(key)
	size, fingerprint := fingerprintOf(data)
	var (
		remembered immutableObservation
		found      bool
	)
	if err := s.withState(func() {
		s.counters.puts++
		remembered, found = s.objects.matching(key, size, &fingerprint)
	}); err != nil {
		return err
	}
	// Bytes this adapter already wrote under this key, unchanged -- but
	// "wrote once" is not "is still there". Collection deletes objects, so
	// the remembered value is confirmed present and unchanged before this
	// reports success. A HEAD is far cheaper than re-sending the payload,
	// which is the point of remembering at all; skipping it would report
	// a durable write for an object that is gone.
	if found {
		current, err := s.immutableObservationIsCurrent(physical, key, remembered)
		if err != nil {
			return err
		}
		if current {
			return s.count(func(c *counters) { c.idempotentPuts++ })
		}
	}

	for attempt := 0; attempt < ConditionalRetryLimit; attempt++ {
		if err := s.count(func(c *counters) {
			c.transportPuts++
			c.transportBytesSubmitted += uint64(len(data))
		}); err != nil {
			return err
		}
		etag, err := s.client.PutObjectIfAbsent(physical, data)
		if err == nil {
			return s.rememberImmutable(key, size, fingerprint, etag)
		}
		if isUnsupported(err) {
			return unsupported("it answered If-None-Match: * with 501 Not Implemented")
		}
		if !isMissing(err) && !isPrecondition(err) {
			return unavailable("immutable put", err)
		}

		// Something is already there. Immutable means the only two
		// answers are "the same bytes, so this write already happened"
		// and "different bytes, which is a real conflict".
		existing, readErr := s.read(physical, key)
		if readErr != nil {
			var typed *Error
			// Deleted between our write losing and our read. Try the
			// whole thing again.
			if errors.As(readErr, &typed) && typed.Code == ErrorMissing {
				continue
			}
			return readErr
		}
		// Remembered either way: a later put of whatever is actually
		// there is then a question the cache can answer, disagreement
		// included.
		foundSize, foundFingerprint := fingerprintOf(existing.Data)
		if err := s.rememberImmutable(
			key, foundSize, foundFingerprint, existing.ETag,
		); err != nil {
			return err
		}
		if !sameBytes(existing.Data, data) {
			return conflict(key)
		}
		return s.count(func(c *counters) { c.idempotentPuts++ })
	}
	return unavailable("immutable put did not settle")
}

// Exists reports whether an immutable object exists.
func (s *S3BucketStorage) Exists(key string) (bool, error) {
	if err := validateKey(key); err != nil {
		return false, err
	}
	if err := s.count(func(c *counters) { c.transportHeads++ }); err != nil {
		return false, err
	}
	found := true
	if _, err := s.client.HeadObject(s.objectKey(key)); err != nil {
		if !isMissing(err) {
			return false, unavailable("head object", err)
		}
		found = false
	}
	if err := s.count(func(c *counters) { c.exists++ }); err != nil {
		return false, err
	}
	return found, nil
}

// List returns the matching logical object keys in lexical order (B4).
func (s *S3BucketStorage) List(prefix string) ([]string, error) {
	if err := validatePrefix(prefix); err != nil {
		return nil, err
	}
	root := s.objectRoot()
	physical := root + prefix
	logical := []string{}
	continuation := ""
	seenTokens := make(map[string]struct{})
	for {
		if err := s.count(func(c *counters) { c.transportLists++ }); err != nil {
			return nil, err
		}
		page, err := s.client.ListObjects(physical, continuation)
		if err != nil {
			return nil, unavailable("list objects", err)
		}
		for _, physicalKey := range page.Keys {
			// A key from outside this adapter's subtree means the client
			// answered about something else. Dropping it quietly would
			// turn a misrouted listing into a short one, which reads as
			// an empty bucket.
			if !strings.HasPrefix(physicalKey, root) {
				return nil, unavailableResponse(
					"a listing returned a key outside the namespace",
				)
			}
			name := strings.TrimPrefix(physicalKey, root)
			// B4: every listed key is one Get accepts, and the prefix
			// filter is the adapter's, not the store's.
			if validateKey(name) != nil || !strings.HasPrefix(name, prefix) {
				continue
			}
			logical = append(logical, name)
		}
		if page.NextToken == "" {
			// More pages promised and none named: a listing that cannot
			// be finished, rather than one that is finished.
			if page.Truncated {
				return nil, unavailableResponse(
					"a listing reported more pages without naming one",
				)
			}
			break
		}
		// A token that repeats is a listing that cannot finish. Refusing
		// beats looping until the process is killed.
		if _, repeated := seenTokens[page.NextToken]; repeated {
			return nil, unavailableResponse(
				"a listing returned an invalid continuation token",
			)
		}
		seenTokens[page.NextToken] = struct{}{}
		continuation = page.NextToken
	}
	// Sorted and de-duplicated here rather than trusted from the client,
	// because pages arrive in the store's order and the seam promises
	// lexical order with no repeats.
	sort.Strings(logical)
	logical = dedupe(logical)
	if err := s.count(func(c *counters) { c.lists++ }); err != nil {
		return nil, err
	}
	return logical, nil
}

func dedupe(sorted []string) []string {
	unique := sorted[:0]
	for index, value := range sorted {
		if index == 0 || value != sorted[index-1] {
			unique = append(unique, value)
		}
	}
	return unique
}

// Delete makes an immutable object logically absent. Deleting what is
// already gone is the state the caller asked for (B14).
func (s *S3BucketStorage) Delete(key string) error {
	if err := validateKey(key); err != nil {
		return err
	}
	if err := s.count(func(c *counters) { c.transportDeletes++ }); err != nil {
		return err
	}
	if err := s.client.DeleteObject(s.objectKey(key)); err != nil && !isMissing(err) {
		return unavailable("delete object", err)
	}
	return s.withState(func() {
		s.objects.forget(key)
		s.counters.deletes++
	})
}

// HeadGet reads the head CAS cell, reporting Missing when it is absent.
func (s *S3BucketStorage) HeadGet() ([]byte, error) { return s.cellGet("head") }

// HeadSwap replaces the head only when its complete value equals
// expected. A nil expected compares against absence (B2).
func (s *S3BucketStorage) HeadSwap(expected, replacement []byte) (bool, error) {
	return s.cellSwap("head", expected, replacement)
}

// FloorGet reads the retention-floor CAS cell, reporting Missing when it
// is absent.
func (s *S3BucketStorage) FloorGet() ([]byte, error) { return s.cellGet("floor") }

// FloorSwap replaces the floor only when its complete value equals
// expected. A nil expected compares against absence (B2).
func (s *S3BucketStorage) FloorSwap(expected, replacement []byte) (bool, error) {
	return s.cellSwap("floor", expected, replacement)
}

// CellPairGet reads head and floor as one finite concurrent observation
// pair, mapping an absent cell to nil (B13).
//
// Two reads, not one observation. The seam allows that: the pair is a
// finite concurrent observation, not an atomic snapshot.
//
// B13 wants the two transport reads overlapped, which is the whole reason
// to offer the capability -- a serial pair is what the caller already has
// in HeadGet then FloorGet. Both goroutines are joined before this
// returns, which is how "completes both calls before returning or
// raising" is kept even when the first read fails.
func (s *S3BucketStorage) CellPairGet() ([]byte, []byte, error) {
	if err := s.withState(func() {}); err != nil {
		return nil, nil, err
	}
	var (
		head, floor       []byte
		headErr, floorErr error
		join              sync.WaitGroup
	)
	read := func(name string, value *[]byte, failure *error) {
		defer join.Done()
		// A panicking client would otherwise take the process down from a
		// goroutine the caller cannot recover in.
		defer func() {
			if panicked := recover(); panicked != nil {
				*value = nil
				*failure = &Error{
					ErrorInternalFailure,
					fmt.Sprintf("S3 bucket %s cell read panicked: %v", name, panicked),
				}
			}
		}()
		*value, *failure = s.optionalCellGet(name)
	}
	join.Add(2)
	go read("head", &head, &headErr)
	go read("floor", &floor, &floorErr)
	join.Wait()
	if headErr != nil {
		return nil, nil, headErr
	}
	if floorErr != nil {
		return nil, nil, floorErr
	}
	return head, floor, nil
}

// Counters returns an observational adapter-counter snapshot. It stays
// readable after Close: a snapshot is an observation, not an operation on
// the store.
func (s *S3BucketStorage) Counters() BucketCounterSnapshot {
	s.mu.Lock()
	defer s.mu.Unlock()
	return BucketCounterSnapshot{
		Gets:                    s.counters.gets,
		Puts:                    s.counters.puts,
		IdempotentPuts:          s.counters.idempotentPuts,
		Exists:                  s.counters.exists,
		Lists:                   s.counters.lists,
		Deletes:                 s.counters.deletes,
		CellGets:                s.counters.cellGets,
		CellSwaps:               s.counters.cellSwaps,
		CellSwapWins:            s.counters.cellSwapWins,
		TransportGets:           s.counters.transportGets,
		TransportPuts:           s.counters.transportPuts,
		TransportHeads:          s.counters.transportHeads,
		TransportLists:          s.counters.transportLists,
		TransportDeletes:        s.counters.transportDeletes,
		TransportBytesReceived:  s.counters.transportBytesReceived,
		TransportBytesSubmitted: s.counters.transportBytesSubmitted,
	}
}
