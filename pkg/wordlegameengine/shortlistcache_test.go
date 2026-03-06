package wordlegameengine

import (
	"sync"
	"testing"
)

func TestMakeCacheKey(t *testing.T) {
	tests := []struct {
		name     string
		guess    string
		feedback string
		wantKey  string
	}{
		{
			name:     "raise with -G---",
			guess:    "raise",
			feedback: "-G---",
			wantKey:  "raise|-G---",
		},
		{
			name:     "crane with GGGGG",
			guess:    "crane",
			feedback: "GGGGG",
			wantKey:  "crane|GGGGG",
		},
		{
			name:     "slate with --G-G",
			guess:    "slate",
			feedback: "--G-G",
			wantKey:  "slate|--G-G",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			guess, err := NewWord(tt.guess)
			if err != nil {
				t.Fatalf("NewWord(%q) error: %v", tt.guess, err)
			}
			feedback, err := ParseFeedback(tt.feedback)
			if err != nil {
				t.Fatalf("ParseFeedback(%q) error: %v", tt.feedback, err)
			}

			key := MakeCacheKey(guess, feedback)
			if string(key) != tt.wantKey {
				t.Errorf("MakeCacheKey() = %q, want %q", key, tt.wantKey)
			}
		})
	}
}

func TestMakeCacheKey_UniqueKeys(t *testing.T) {
	// Test that different (guess, feedback) pairs create unique keys
	keys := make(map[CacheKey]bool)

	guess1, _ := NewWord("raise")
	feedback1, _ := ParseFeedback("-G---")
	key1 := MakeCacheKey(guess1, feedback1)
	keys[key1] = true

	guess2, _ := NewWord("raise")
	feedback2, _ := ParseFeedback("GGGGG")
	key2 := MakeCacheKey(guess2, feedback2)
	if keys[key2] {
		t.Error("Different feedback should produce different key")
	}
	keys[key2] = true

	guess3, _ := NewWord("crane")
	feedback3, _ := ParseFeedback("-G---")
	key3 := MakeCacheKey(guess3, feedback3)
	if keys[key3] {
		t.Error("Different guess should produce different key")
	}
}

func TestShortlistCache_PutAndGet(t *testing.T) {
	cache := NewShortlistCache()

	// Create a test shortlist
	word1, _ := NewWord("apple")
	word2, _ := NewWord("crane")
	word3, _ := NewWord("raise")
	shortlist := []Word{word1, word2, word3}

	// Create a cache key
	guess, _ := NewWord("raise")
	feedback, _ := ParseFeedback("-G---")
	key := MakeCacheKey(guess, feedback)

	// Put into cache
	cache.Put(key, shortlist)

	// Get from cache
	retrieved, found := cache.Get(key)
	if !found {
		t.Fatal("Expected to find key in cache")
	}

	// Verify the retrieved shortlist matches
	if len(retrieved) != len(shortlist) {
		t.Errorf("Retrieved shortlist length = %d, want %d", len(retrieved), len(shortlist))
	}
	for i, word := range shortlist {
		if retrieved[i] != word {
			t.Errorf("Retrieved shortlist[%d] = %v, want %v", i, retrieved[i], word)
		}
	}
}

func TestShortlistCache_Get_NotFound(t *testing.T) {
	cache := NewShortlistCache()

	guess, _ := NewWord("raise")
	feedback, _ := ParseFeedback("-G---")
	key := MakeCacheKey(guess, feedback)

	_, found := cache.Get(key)
	if found {
		t.Error("Expected not to find key in empty cache")
	}
}

func TestShortlistCache_ModifyReturnedShortlist(t *testing.T) {
	cache := NewShortlistCache()

	// Create a test shortlist
	word1, _ := NewWord("apple")
	word2, _ := NewWord("crane")
	shortlist := []Word{word1, word2}

	guess, _ := NewWord("raise")
	feedback, _ := ParseFeedback("-G---")
	key := MakeCacheKey(guess, feedback)

	// Put into cache
	cache.Put(key, shortlist)

	// Get from cache and modify
	retrieved, _ := cache.Get(key)
	word3, _ := NewWord("slate")
	retrieved = append(retrieved, word3)

	// Get again and verify original is unchanged
	retrieved2, _ := cache.Get(key)
	if len(retrieved2) != 2 {
		t.Errorf("Cache shortlist was modified by external code: length = %d, want 2", len(retrieved2))
	}
}

func TestShortlistCache_ThreadSafety(t *testing.T) {
	cache := NewShortlistCache()

	// Run concurrent writes and reads
	var wg sync.WaitGroup
	numGoroutines := 100

	// Concurrent writes
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			guess, _ := NewWord("raise")
			feedback, _ := ParseFeedback("-G---")
			key := MakeCacheKey(guess, feedback)
			word, _ := NewWord("apple")
			cache.Put(key, []Word{word})
		}(i)
	}

	// Concurrent reads
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			guess, _ := NewWord("raise")
			feedback, _ := ParseFeedback("-G---")
			key := MakeCacheKey(guess, feedback)
			cache.Get(key)
		}(i)
	}

	wg.Wait()

	// Verify cache is still functional
	guess, _ := NewWord("raise")
	feedback, _ := ParseFeedback("-G---")
	key := MakeCacheKey(guess, feedback)
	_, found := cache.Get(key)
	if !found {
		t.Error("Cache should have entry after concurrent operations")
	}
}

func TestShortlistCache_ReplaceOrInsert(t *testing.T) {
	cache := NewShortlistCache()

	guess, _ := NewWord("raise")
	feedback, _ := ParseFeedback("-G---")
	key := MakeCacheKey(guess, feedback)

	word1, _ := NewWord("apple")
	word2, _ := NewWord("crane")

	// First insert
	cache.Put(key, []Word{word1})

	// Replace with new value
	cache.Put(key, []Word{word2})

	// Verify replacement
	retrieved, _ := cache.Get(key)
	if len(retrieved) != 1 || retrieved[0] != word2 {
		t.Errorf("Expected replacement to work, got %v", retrieved)
	}
}

func TestCacheEntry_Less(t *testing.T) {
	tests := []struct {
		name     string
		a        CacheEntry
		b        CacheEntry
		expected bool
	}{
		{
			name:     "a < b alphabetically",
			a:        CacheEntry{Key: "apple|GGGGG"},
			b:        CacheEntry{Key: "crane|GGGGG"},
			expected: true,
		},
		{
			name:     "a > b alphabetically",
			a:        CacheEntry{Key: "crane|GGGGG"},
			b:        CacheEntry{Key: "apple|GGGGG"},
			expected: false,
		},
		{
			name:     "a == b",
			a:        CacheEntry{Key: "raise|-G---"},
			b:        CacheEntry{Key: "raise|-G---"},
			expected: false,
		},
		{
			name:     "different feedback for same guess",
			a:        CacheEntry{Key: "raise|-G---"},
			b:        CacheEntry{Key: "raise|GGGGG"},
			expected: true, // '-' < 'G'
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.a.Less(tt.b)
			if result != tt.expected {
				t.Errorf("Less() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestInitCache(t *testing.T) {
	// Save any existing cache
	oldCache := FirstTurnCache
	defer func() {
		FirstTurnCache = oldCache
	}()

	// Initialize cache
	InitCache()

	if FirstTurnCache == nil {
		t.Error("InitCache() should initialize FirstTurnCache")
	}

	// Verify it's functional
	guess, _ := NewWord("raise")
	feedback, _ := ParseFeedback("-G---")
	key := MakeCacheKey(guess, feedback)
	word, _ := NewWord("apple")

	FirstTurnCache.Put(key, []Word{word})
	retrieved, found := FirstTurnCache.Get(key)
	if !found || len(retrieved) != 1 {
		t.Error("Global cache should be functional after InitCache()")
	}
}
