package analysis

import (
	"testing"
	"time"
)

func TestInMemoryCache_SetGet(t *testing.T) {
	c := NewInMemoryCache()

	report := &AnalysisReport{ID: "test-1", OverallScore: 0.85}
	c.Set("key1", report, 10*time.Minute)

	got, ok := c.Get("key1")
	if !ok {
		t.Fatal("expected cache hit")
	}
	if got.ID != "test-1" {
		t.Fatalf("got ID %q, want %q", got.ID, "test-1")
	}
}

func TestInMemoryCache_Miss(t *testing.T) {
	c := NewInMemoryCache()

	_, ok := c.Get("nonexistent")
	if ok {
		t.Fatal("expected cache miss")
	}
}

func TestInMemoryCache_Expiration(t *testing.T) {
	c := NewInMemoryCache()

	report := &AnalysisReport{ID: "expired"}
	c.Set("key1", report, 1*time.Millisecond)

	// Wait for expiration
	time.Sleep(5 * time.Millisecond)

	_, ok := c.Get("key1")
	if ok {
		t.Fatal("expected cache miss for expired entry")
	}
}

func TestInMemoryCache_Delete(t *testing.T) {
	c := NewInMemoryCache()
	c.Set("key1", &AnalysisReport{ID: "del"}, 10*time.Minute)

	c.Delete("key1")

	_, ok := c.Get("key1")
	if ok {
		t.Fatal("expected cache miss after delete")
	}
}

func TestInMemoryCache_Clear(t *testing.T) {
	c := NewInMemoryCache()
	c.Set("a", &AnalysisReport{}, 10*time.Minute)
	c.Set("b", &AnalysisReport{}, 10*time.Minute)

	c.Clear()

	if c.Size() != 0 {
		t.Fatalf("got size %d after clear, want 0", c.Size())
	}
}

func TestInMemoryCache_MaxSize(t *testing.T) {
	c := NewInMemoryCache(WithMaxSize(2))

	c.Set("first", &AnalysisReport{ID: "first"}, 10*time.Minute)
	time.Sleep(1 * time.Millisecond)
	c.Set("second", &AnalysisReport{ID: "second"}, 10*time.Minute)
	time.Sleep(1 * time.Millisecond)
	c.Set("third", &AnalysisReport{ID: "third"}, 10*time.Minute)

	if c.Size() != 2 {
		t.Fatalf("got size %d, want 2", c.Size())
	}

	// Oldest entry ("first") should be evicted
	_, ok := c.Get("first")
	if ok {
		t.Fatal("expected oldest entry to be evicted")
	}

	// Second and third should still be there
	if _, ok := c.Get("second"); !ok {
		t.Fatal("expected 'second' to still be in cache")
	}
	if _, ok := c.Get("third"); !ok {
		t.Fatal("expected 'third' to still be in cache")
	}
}

func TestInMemoryCache_Stats(t *testing.T) {
	c := NewInMemoryCache(WithMaxSize(10))

	c.Set("hit", &AnalysisReport{}, 10*time.Minute)
	c.Get("hit")   // hit
	c.Get("miss1") // miss
	c.Get("miss2") // miss

	stats := c.Stats()
	if stats.Hits != 1 {
		t.Fatalf("got %d hits, want 1", stats.Hits)
	}
	if stats.Misses != 2 {
		t.Fatalf("got %d misses, want 2", stats.Misses)
	}
	if stats.Size != 1 {
		t.Fatalf("got size %d, want 1", stats.Size)
	}
	if stats.MaxSize != 10 {
		t.Fatalf("got maxSize %d, want 10", stats.MaxSize)
	}
	// Hit rate should be 1/3 ≈ 0.333
	if stats.HitRate < 0.3 || stats.HitRate > 0.4 {
		t.Fatalf("got hitRate %f, want ~0.333", stats.HitRate)
	}
}

func TestInMemoryCache_Prune(t *testing.T) {
	c := NewInMemoryCache()

	c.Set("fresh", &AnalysisReport{}, 10*time.Minute)
	c.Set("stale", &AnalysisReport{}, 1*time.Millisecond)
	time.Sleep(5 * time.Millisecond)

	pruned := c.Prune()
	if pruned != 1 {
		t.Fatalf("pruned %d, want 1", pruned)
	}
	if c.Size() != 1 {
		t.Fatalf("got size %d, want 1 after prune", c.Size())
	}
}

func TestCacheKey(t *testing.T) {
	k1 := CacheKey("git", "https://github.com/example/repo")
	k2 := CacheKey("git", "https://github.com/example/repo")
	k3 := CacheKey("web", "https://github.com/example/repo")

	if k1 != k2 {
		t.Fatal("same inputs should produce same key")
	}
	if k1 == k3 {
		t.Fatal("different source types should produce different keys")
	}
}

func TestNullCache(t *testing.T) {
	c := NullCache{}

	c.Set("key", &AnalysisReport{}, time.Hour)
	_, ok := c.Get("key")
	if ok {
		t.Fatal("NullCache should always miss")
	}
	c.Delete("key")
	c.Clear()

	if c.Size() != 0 {
		t.Fatal("NullCache size should be 0")
	}

	stats := c.Stats()
	if stats.Hits != 0 || stats.Misses != 0 {
		t.Fatal("NullCache stats should be zero")
	}
}

func TestCacheEntry_IsExpired(t *testing.T) {
	e := CacheEntry{
		ExpiresAt: time.Now().Add(-time.Second),
	}
	if !e.IsExpired() {
		t.Fatal("entry should be expired")
	}

	e2 := CacheEntry{
		ExpiresAt: time.Now().Add(time.Hour),
	}
	if e2.IsExpired() {
		t.Fatal("entry should not be expired")
	}
}
