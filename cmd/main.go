package main

import (
	"cache-system/internal"
	"fmt"
	"time"
)

func main() {
	// create a cache
	cacheSystem := internal.NewCache()
	// add key-value pairs to it
	cacheSystem.Set("age", 24, 50*time.Second)
	// Retrieve values
	key := "age"
	val, ok := cacheSystem.Get(key)
	if ok == true {
		fmt.Printf("key: %s \tval: %v\n", key, val)
	}
	key = "name"
	val, ok = cacheSystem.Get(key)
	if ok == true {
		fmt.Printf("key: %s \tval: %v\n", key, val)
	}
}