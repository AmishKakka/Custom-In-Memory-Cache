package internal

import (
	"fmt"
	"os"
	"time"
)

// Creates a new instance of Cache
func NewCache() *cache {
	c := &cache{
		entry: make(map[string]*entry),
	}
	// creating the append-only file
	file, err := os.OpenFile("server.wal", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic(fmt.Errorf("failed to open write-ahead log: %w", err))
	}
	c.walFile = file
	return c
}

// Aquires the lock and creates/updates key
func (c *cache) set(key string, val any, ttl time.Duration) {
	c.cacheMU.Lock()
	defer c.cacheMU.Unlock()
	// core operation, setting val & TTL
	expiration := time.Now().Add(ttl)
	c.entry[key] = &entry{
		Value: val,
		TTL: expiration,
	}
	// fmt.Printf("key: %s \tval: %v\n", key, val)
}

// WAL and then SET updates to Cache
func (c *cache) SET(key string, val any, ttl time.Duration) {
	// acquiring the WAL file
	c.walMU.Lock()
	// write to the file
	expiration := time.Now().Add(ttl)
	_, err := fmt.Fprintf(c.walFile, "SET|%s|%v|%d\n", key, val, expiration.UnixNano())
	if err == nil {
		// writing to disk the file contents
		c.walFile.Sync()
	}
	c.walMU.Unlock()
	// Now, call set() to make chanegs to Cache
	c.set(key, val, ttl)
}

// Retrieve the key from cache
func (c *cache) GET(key string) (any, bool) {
	c.cacheMU.RLock()
	defer c.cacheMU.RUnlock()
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
func (c *cache) delete(key string) {
	c.cacheMU.Lock()
	defer c.cacheMU.Unlock()
	// No need to check if key exists or is expired, 
	// Go’s built-in delete() function is safe to call even if the key isn't in the map—it will simply do nothing and return.
	delete(c.entry, key)
}

// WAL and then DEL updates to cache
func (c *cache) DEL(key string) {
	// acquiring the WAL file
	c.walMU.Lock()
	// write to the file
	_, err := fmt.Fprintf(c.walFile, "DEL|%s\n", key)
	if err == nil {
		// writing to disk the file contents
		c.walFile.Sync()
	}
	c.walMU.Unlock()
	// Now, call set() to make chanegs to Cache
	c.delete(key)
}

// Clean up cache
func (c *cache) Cleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	// iterate over the channel where ticks are delivered
	for range ticker.C {
		c.walMU.Lock()
		c.cacheMU.Lock()
		// Removing expired keys
		for k, v := range c.entry {
			if time.Now().After(v.TTL) {
				_, err := fmt.Fprintf(c.walFile, "DEL|%s\n", k)
				if err == nil {
					// writing to disk the file contents
					c.walFile.Sync()
				}
				delete(c.entry, k)
			}
		}
		// will free the lock as soon as this loop finishes and not when the function finishes
		c.cacheMU.Unlock()
		c.walMU.Unlock()
	}
}