package internal

import (
	// "fmt"
	"time"
)

// Creates a new instance of Cache
func NewCache() *cache {
	return &cache{
		entry: make(map[string]*entry),
	}
}

// Aquires the lock and creates/updates key
func (c *cache) Set(key string, val any, ttl time.Duration) {
	c.Lock()
	defer c.Unlock()
	// core operation, setting val & TTL
	expiration := time.Now().Add(ttl)
	c.entry[key] = &entry{
		Value: val,
		TTL: expiration,
	}
	// fmt.Printf("key: %s \tval: %v\n", key, val)
}

// Retrieve the key from cache
func (c *cache) Get(key string) (any, bool) {
	c.RLock()
	defer c.RUnlock()
	// see if key exists
	item, exists := c.entry[key]
	if exists == false {
		// fmt.Printf("'%s' key does not exist.\n", key)
		return nil, false
	}
	if time.Now().After(item.TTL) {
		// fmt.Printf("'%s' key has expired.\n", key)
		return nil, false
	}
	return item.Value, true
}

// Delete a key if present in the cache
func (c *cache) Delete(key string) bool {
	c.Lock()
	defer c.Unlock()
	// No need to check if key exists or is expired, 
	// Go’s built-in delete() function is safe to call even if the key isn't in the map—it will simply do nothing and return.
	delete(c.entry, key)
	return true
}

// Clean up cache
func (c *cache) Cleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	// iterate over the channel where ticks are delivered
	for range ticker.C {
		c.Lock()
		// Removing expired keys
		for k, v := range c.entry {
			if time.Now().After(v.TTL) {
				delete(c.entry, k)
			}
		}
		// will free the lock as soon as this loop finishes and not when the function finishes
		c.Unlock()
	}
}