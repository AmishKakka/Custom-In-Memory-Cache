package internal

import (
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
	sync.RWMutex
	entry map[string]*entry
}