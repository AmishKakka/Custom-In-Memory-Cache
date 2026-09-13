package main

import (
	"cache-system/internal"
	"fmt"
	"time"
	"sync"
	"math/rand"
)

func main() {
	// create a cache
	cache := internal.NewCache()
	// // add key-value pairs to it
	// cache.Set("age", 24, 50*time.Second)
	// // Retrieve values
	// key := "age"
	// val, ok := cache.Get(key)
	// if ok == true {
	// 	fmt.Printf("key: %s \tval: %v\n", key, val)
	// }
	// key = "name"
	// val, ok = cache.Get(key)
	// if ok == true {
	// 	fmt.Printf("key: %s \tval: %v\n", key, val)
	// }

	go cache.Cleanup(50 * time.Millisecond)

	fmt.Println("Running 5 workers x 6000 iterations (180,000 total cache operations)...")

	var wg sync.WaitGroup
	workersCount := 5
	iterationsPerWorker := 6000

	// Launching the concurrent workers
	for i := 1; i <= workersCount; i++ {
		wg.Add(1)
		go func(workerID int) {
			// when thsi function finishes 'defer' will be called
			defer wg.Done()
			// Seed local pseudo-random generator uniquely per worker
			// to avoid global rand lock bottlenecks in high concurrency
			r := rand.New(rand.NewSource(time.Now().UnixNano() + int64(workerID)))

			for j := 0; j < iterationsPerWorker; j++ {
				// We pick random key numbers from 0 to 99 to ensure keys collide across workers.
				// For lookups, we occasionally look up keys 100-110 to simulate keys that don't exist.
				randomSetKey := fmt.Sprintf("account_%d", r.Intn(20))
				randomGetKey := fmt.Sprintf("account_%d", r.Intn(30)) 
				randomDelKey := fmt.Sprintf("account_%d", r.Intn(15))

				// Set a dynamic key with varying values and 40ms TTL
				randomVal := r.Float64() * 5000
				cache.Set(randomSetKey, randomVal, 40*time.Millisecond)

				// Get a key (may exist, may not exist, or might have expired)
				cache.Get(randomGetKey)

				// Set value to a heavily used key
				cache.Set("master_ledger", r.Intn(99999), 100*time.Millisecond)

				// read the key
				cache.Get(randomGetKey)

				// Delete a random key
				cache.Delete(randomDelKey)

				// again set a vlaue to random key
				cache.Get(randomSetKey)
			}
			fmt.Printf("Worker %d successfully processed %d iterations (180k ops)\n", workerID, iterationsPerWorker)
		}(i)
	}
	// block the main execution thread until all 5 workers finish their tasks
	startTime := time.Now()
	wg.Wait()
	duration := time.Since(startTime)
	fmt.Printf("\nAll concurrent operations completed cleanly in %v!\n", duration)

	// final verification of the background cleanup process
	fmt.Println("\nVerifying final background state...")
	
	// adding fresh keys
	cache.Set("final_check_1", "alive", 100*time.Millisecond)
	cache.Set("final_check_2", "expired_soon", 10*time.Millisecond)

	// Let the cleanup and let keys naturally pass their TTL
	time.Sleep(200 * time.Millisecond)
	_, ok1 := cache.Get("final_check_1")
	_, ok2 := cache.Get("final_check_2")

	fmt.Printf("Key 1 active check (Expected false): %v\n", ok1)
	fmt.Printf("Key 2 active check (Expected false): %v\n", ok2)
}