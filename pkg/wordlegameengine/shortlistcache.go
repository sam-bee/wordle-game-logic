package wordlegameengine

import (
	"fmt"
	"sync"

	"github.com/google/btree"
)

// CacheKey is a string in format "guess|feedback" (e.g., "raise|-G---")
type CacheKey string

// MakeCacheKey creates a cache key from a guess and feedback
func MakeCacheKey(guess Word, feedback Feedback) CacheKey {
	return CacheKey(fmt.Sprintf("%s|%s", guess.String(), feedback.String()))
}

// CacheEntry implements btree.Item interface
type CacheEntry struct {
	Key       CacheKey
	Shortlist []Word // Copy of the shortlist
}

// Less implements btree.Item for ordering
func (e CacheEntry) Less(than btree.Item) bool {
	other, ok := than.(CacheEntry)
	if !ok {
		return false
	}
	return e.Key < other.Key
}

// ShortlistCache is a thread-safe B-tree cache
type ShortlistCache struct {
	tree   *btree.BTree
	mutex  sync.RWMutex
	degree int
}

const BTreeDegree = 32

// NewShortlistCache creates a new cache with BTreeDegree = 32
func NewShortlistCache() *ShortlistCache {
	return &ShortlistCache{
		tree:   btree.New(BTreeDegree),
		degree: BTreeDegree,
	}
}

// Get retrieves a shortlist from cache (thread-safe)
func (c *ShortlistCache) Get(key CacheKey) ([]Word, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	item := c.tree.Get(CacheEntry{Key: key})
	if item == nil {
		return nil, false
	}

	entry, ok := item.(CacheEntry)
	if !ok {
		return nil, false
	}

	// Return a copy to prevent external modification
	shortlistCopy := make([]Word, len(entry.Shortlist))
	copy(shortlistCopy, entry.Shortlist)
	return shortlistCopy, true
}

// Put stores a shortlist in cache (thread-safe, makes copy of shortlist)
func (c *ShortlistCache) Put(key CacheKey, shortlist []Word) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Make a copy of the shortlist to store
	shortlistCopy := make([]Word, len(shortlist))
	copy(shortlistCopy, shortlist)

	entry := CacheEntry{
		Key:       key,
		Shortlist: shortlistCopy,
	}
	c.tree.ReplaceOrInsert(entry)
}

// Global cache instance
var FirstTurnCache *ShortlistCache

// InitCache initializes the global cache
func InitCache() {
	FirstTurnCache = NewShortlistCache()
}
