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
	// fixed length channel
	c.walEntry = make(chan string, 1024)
	return c
}

// Consumes items from WAL channel and writes to log file
func (c *cache) WriteToFile() {
	const batchSize = 256
	count := 0
	for entry := range c.walEntry {
		// adding entry to log file
		fmt.Fprint(c.walFile, entry)
		count += 1
		if count >= batchSize {
			c.walFile.Sync()
			// fmt.Printf("%v items added to log file.", count)
			/* 
				by doing this, the method once consumed items will wait 
				if the WAL channel is empty or it has added 'batchSize' mutations to the log file.
				It will not block any other processes.
			*/
			count = 0
		}
	}
	if count > 0 {
		c.walFile.Sync()
	}
}

// Aquires the lock and creates/updates key
func (c *cache) set(key string, val any, expiration time.Time) {
	c.cacheMU.Lock()
	defer c.cacheMU.Unlock()
	// core operation, setting val & TTL
	c.entry[key] = &entry{
		Value: val,
		TTL: expiration,
	}
	// fmt.Printf("key: %s \tval: %v\n", key, val)
}

// WAL and then SET updates to Cache
func (c *cache) SET(key string, val any, ttl time.Duration) {
	expiration := time.Now().Add(ttl)
	// creating the 'entry' string and adding to the WAL channel
	entry := fmt.Sprintf("S|%s|%v|%d\n", key, val, expiration.UnixNano())
	c.walEntry <- entry
	// Now, call set() to make chanegs to Cache
	c.set(key, val, expiration)
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
	// creating the 'entry' string and adding to the WAL channel
	entry := fmt.Sprintf("D|%s\n", key)
	c.walEntry <- entry
	// Now, call set() to make chanegs to Cache
	c.delete(key)
}

// Clean up cache
func (c *cache) Cleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	// iterate over the channel where ticks are delivered
	for range ticker.C {
		var expiredKeys []string
		c.cacheMU.Lock()
		// Removing expired keys
		for k, v := range c.entry {
			if time.Now().After(v.TTL) {
				expiredKeys = append(expiredKeys, k)
				delete(c.entry, k)
			}
		}
		// will free the lock
		c.cacheMU.Unlock()
		// move the deleted items entry to WAL channel
		for _, k :=  range expiredKeys {
			c.walEntry <- fmt.Sprintf("D|%s\n", k)
		}
	}
}