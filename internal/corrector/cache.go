package corrector

import (
	"container/list"
	"sync"
)

// ResultCache is a simple thread-safe LRU cache for correction results.
type ResultCache struct {
	mu       sync.RWMutex
	capacity int
	items    map[string]*list.Element
	order    *list.List
}

type cacheEntry struct {
	key    string
	result CorrectionResult
}

// NewResultCache creates a new LRU cache with the given capacity.
func NewResultCache(capacity int) *ResultCache {
	if capacity <= 0 {
		capacity = 128
	}
	return &ResultCache{
		capacity: capacity,
		items:    make(map[string]*list.Element),
		order:    list.New(),
	}
}

// Get retrieves a cached result if present.
func (c *ResultCache) Get(key string) (CorrectionResult, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	el, ok := c.items[key]
	if !ok {
		return CorrectionResult{}, false
	}
	c.order.MoveToFront(el)
	return el.Value.(*cacheEntry).result, true
}

// Set stores a result in the cache.
func (c *ResultCache) Set(key string, result CorrectionResult) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if el, ok := c.items[key]; ok {
		c.order.MoveToFront(el)
		el.Value.(*cacheEntry).result = result
		return
	}

	el := c.order.PushFront(&cacheEntry{key: key, result: result})
	c.items[key] = el

	if c.order.Len() > c.capacity {
		back := c.order.Back()
		if back != nil {
			c.order.Remove(back)
			delete(c.items, back.Value.(*cacheEntry).key)
		}
	}
}

// Len returns the current number of cached items.
func (c *ResultCache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.order.Len()
}

// Clear empties the cache.
func (c *ResultCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[string]*list.Element)
	c.order = list.New()
}
