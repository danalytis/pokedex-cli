package pokecache

import (
	"sync"
	"time"
)

type cacheEntry struct {
	createdAt time.Time
	val       []byte
}

type Cache struct {
	mu         sync.Mutex
	cacheEntry map[string]cacheEntry
	interval   time.Duration
}

func (c *Cache) Add(key string, val []byte) {
	if key == "" || val == nil || len(val) == 0 {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	newEntry := cacheEntry{}
	newEntry.createdAt = time.Now()
	newEntry.val = val
	c.cacheEntry[key] = newEntry
}

func (c *Cache) Get(key string) ([]byte, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	value, ok := c.cacheEntry[key]
	if ok {
		return value.val, true
	} else {
		return nil, false
	}
}

func (c *Cache) reapLoop(ticker *time.Ticker) {
	for range ticker.C {
		c.mu.Lock()
		defer c.mu.Unlock()

		for url, entry := range c.cacheEntry {
			cutoff := time.Now().Add(-c.interval)
			if entry.createdAt.Before(cutoff) {
				delete(c.cacheEntry, url)
			}
		}

	}
}

func NewCache(interval time.Duration) Cache {
	c := Cache{}
	c.cacheEntry = make(map[string]cacheEntry)
	c.interval = interval
	t := time.NewTicker(interval)
	go c.reapLoop(t)
	return c
}
