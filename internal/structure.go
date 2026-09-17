package internal

import (
	"container/list"
	"os"
	"sync"
	"time"
)

// A single entry in cache
type entry struct {
	Key string
	Value any
	TTL time.Time
}

// Cache sytem with Read/Write mutex called anonymously
type cache struct {
	// Mutex exclusively for the map
	cacheMU sync.RWMutex
	entry map[string]*list.Element
	ll *list.List
	// size of the entire cache
	maxSize int
	// removed lock on file because introduced channels
	walFile *os.File
	// creating a channel to store all mutations and then consume fixed no. of items
	walEntry chan string
}