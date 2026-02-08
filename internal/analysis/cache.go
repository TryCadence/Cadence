package analysis

import (
	"crypto/sha256"
	"fmt"
	"sync"
	"time"
)

// CacheEntry holds a cached analysis result with expiration metadata.
type CacheEntry struct {
	Report    *AnalysisReport
	CreatedAt time.Time
	ExpiresAt time.Time
}

// IsExpired returns true if the cache entry has passed its TTL.
func (ce *CacheEntry) IsExpired() bool {
	return time.Now().After(ce.ExpiresAt)
}

// AnalysisCache defines the interface for caching analysis results.
type AnalysisCache interface {
	// Get retrieves a cached report by key. Returns nil, false on miss.
	Get(key string) (*AnalysisReport, bool)

	// Set stores a report with a TTL.
	Set(key string, report *AnalysisReport, ttl time.Duration)

	// Delete removes a specific entry.
	Delete(key string)

	// Clear removes all entries.
	Clear()

	// Size returns the number of cached entries (including expired).
	Size() int

	// Stats returns cache statistics.
	Stats() CacheStats
}

// CacheStats provides visibility into cache performance.
type CacheStats struct {
	Hits      int64   `json:"hits"`
	Misses    int64   `json:"misses"`
	Evictions int64   `json:"evictions"`
	Size      int     `json:"size"`
	MaxSize   int     `json:"maxSize"`
	HitRate   float64 `json:"hitRate"`
}

// InMemoryCache is a thread-safe in-memory implementation of AnalysisCache
// with TTL-based expiration and an optional max-size eviction policy (LRU-ish).
type InMemoryCache struct {
	mu        sync.RWMutex
	entries   map[string]*CacheEntry
	maxSize   int
	hits      int64
	misses    int64
	evictions int64
}

// CacheOption configures an InMemoryCache.
type CacheOption func(*InMemoryCache)

// WithMaxSize sets the maximum number of entries the cache will hold.
// When the limit is exceeded the oldest entry is evicted.
// A value of 0 means unlimited.
func WithMaxSize(n int) CacheOption {
	return func(c *InMemoryCache) {
		c.maxSize = n
	}
}

// NewInMemoryCache creates a new in-memory cache.
func NewInMemoryCache(opts ...CacheOption) *InMemoryCache {
	c := &InMemoryCache{
		entries: make(map[string]*CacheEntry),
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Get retrieves a cached report. Expired entries are treated as misses and removed.
func (c *InMemoryCache) Get(key string) (*AnalysisReport, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, ok := c.entries[key]
	if !ok {
		c.misses++
		return nil, false
	}

	if entry.IsExpired() {
		delete(c.entries, key)
		c.misses++
		c.evictions++
		return nil, false
	}

	c.hits++
	return entry.Report, true
}

// Set stores a report with the given TTL.
func (c *InMemoryCache) Set(key string, report *AnalysisReport, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	c.entries[key] = &CacheEntry{
		Report:    report,
		CreatedAt: now,
		ExpiresAt: now.Add(ttl),
	}

	// Enforce max size — evict oldest entry if necessary
	if c.maxSize > 0 && len(c.entries) > c.maxSize {
		c.evictOldestLocked()
	}
}

// Delete removes a specific entry.
func (c *InMemoryCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.entries, key)
}

// Clear removes all entries and resets counters.
func (c *InMemoryCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = make(map[string]*CacheEntry)
	c.hits = 0
	c.misses = 0
	c.evictions = 0
}

// Size returns the current number of entries (may include expired ones).
func (c *InMemoryCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.entries)
}

// Stats returns current cache statistics.
func (c *InMemoryCache) Stats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	total := c.hits + c.misses
	var hitRate float64
	if total > 0 {
		hitRate = float64(c.hits) / float64(total)
	}

	return CacheStats{
		Hits:      c.hits,
		Misses:    c.misses,
		Evictions: c.evictions,
		Size:      len(c.entries),
		MaxSize:   c.maxSize,
		HitRate:   hitRate,
	}
}

// Prune removes all expired entries and returns the number removed.
func (c *InMemoryCache) Prune() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	pruned := 0
	for key, entry := range c.entries {
		if entry.IsExpired() {
			delete(c.entries, key)
			pruned++
			c.evictions++
		}
	}
	return pruned
}

// evictOldestLocked removes the entry with the earliest creation time.
// Caller must hold the write lock.
func (c *InMemoryCache) evictOldestLocked() {
	var oldestKey string
	var oldestTime time.Time

	for key, entry := range c.entries {
		if oldestKey == "" || entry.CreatedAt.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.CreatedAt
		}
	}

	if oldestKey != "" {
		delete(c.entries, oldestKey)
		c.evictions++
	}
}

// CacheKey generates a deterministic cache key for a given source type and identifier.
func CacheKey(sourceType, identifier string) string {
	h := sha256.Sum256([]byte(fmt.Sprintf("%s:%s", sourceType, identifier)))
	return fmt.Sprintf("%s:%x", sourceType, h[:8])
}

// NullCache is a no-op implementation of AnalysisCache for use when caching is disabled.
type NullCache struct{}

func (NullCache) Get(string) (*AnalysisReport, bool)         { return nil, false }
func (NullCache) Set(string, *AnalysisReport, time.Duration) {}
func (NullCache) Delete(string)                              {}
func (NullCache) Clear()                                     {}
func (NullCache) Size() int                                  { return 0 }
func (NullCache) Stats() CacheStats                          { return CacheStats{} }
