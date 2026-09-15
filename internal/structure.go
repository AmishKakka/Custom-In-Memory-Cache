package internal

import (
	"os"
	"sync"
	"time"
)

// A single entry in cache
type entry struct {
	Value any
	TTL time.Time
}

// Cache sytem with Read/Write mutex called anonymously
type cache struct {
	// Mutex exclusively for the map
	cacheMU sync.RWMutex
	entry map[string]*entry
	// mutex exclusively for the append-only file
	walMU sync.Mutex
	walFile *os.File
}