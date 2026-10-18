package corrector

import (
	"testing"
)

func TestResultCacheBasic(t *testing.T) {
	cache := NewResultCache(3)

	r1 := CorrectionResult{Original: "a", Corrected: "b", Confidence: 0.9}
	cache.Set("a", r1)

	got, ok := cache.Get("a")
	if !ok {
		t.Fatal("Expected cache hit for 'a'")
	}
	if got.Corrected != "b" {
		t.Errorf("Expected corrected='b', got %q", got.Corrected)
	}
	if cache.Len() != 1 {
		t.Errorf("Expected len=1, got %d", cache.Len())
	}
}

func TestResultCacheLRUEviction(t *testing.T) {
	cache := NewResultCache(2)
	cache.Set("a", CorrectionResult{Original: "a"})
	cache.Set("b", CorrectionResult{Original: "b"})
	cache.Set("c", CorrectionResult{Original: "c"})

	if cache.Len() != 2 {
		t.Errorf("Expected len=2 after eviction, got %d", cache.Len())
	}

	if _, ok := cache.Get("a"); ok {
		t.Error("Expected 'a' to be evicted")
	}
	if _, ok := cache.Get("b"); !ok {
		t.Error("Expected 'b' to still be present")
	}
	if _, ok := cache.Get("c"); !ok {
		t.Error("Expected 'c' to be present")
	}
}

func TestResultCacheUpdateExisting(t *testing.T) {
	cache := NewResultCache(2)
	cache.Set("a", CorrectionResult{Original: "a", Corrected: "old"})
	cache.Set("a", CorrectionResult{Original: "a", Corrected: "new"})

	got, ok := cache.Get("a")
	if !ok {
		t.Fatal("Expected cache hit")
	}
	if got.Corrected != "new" {
		t.Errorf("Expected updated value 'new', got %q", got.Corrected)
	}
}

func TestResultCacheClear(t *testing.T) {
	cache := NewResultCache(10)
	cache.Set("a", CorrectionResult{Original: "a"})
	cache.Clear()
	if cache.Len() != 0 {
		t.Errorf("Expected len=0 after clear, got %d", cache.Len())
	}
	if _, ok := cache.Get("a"); ok {
		t.Error("Expected cache miss after clear")
	}
}

func TestCorrectorCacheIntegration(t *testing.T) {
	c, err := NewCorrector()
	if err != nil {
		t.Fatalf("NewCorrector: %v", err)
	}

	// First call should populate cache
	_ = c.Correct("电恼")
	if c.CacheStats() == 0 {
		t.Error("Expected cache to have entries after Correct()")
	}

	// Disable cache
	c.DisableCache()
	before := c.CacheStats()
	_ = c.Correct("手机")
	if c.CacheStats() != before {
		t.Error("Expected cache size unchanged when disabled")
	}

	// Re-enable
	c.EnableCache()
	_ = c.Correct("手机")
	if c.CacheStats() <= before {
		t.Error("Expected cache size to grow after re-enabling")
	}
}
