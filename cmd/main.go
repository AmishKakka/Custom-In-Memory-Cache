package main

import (
	"cache-system/internal"
	"fmt"
	"time"
	"sync"
	"math/rand"
	"runtime"
)

func printMemoryStats(label string) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	fmt.Printf("\nMemory Stats [%s]:\n", label)
	fmt.Printf("Heap Alloc:      %0.2f MB (Active data in RAM)\n", float64(m.Alloc)/1024/1024)
	fmt.Printf("Total Alloc:     %0.2f MB (Cumulative lifetime memory churn)\n", float64(m.TotalAlloc)/1024/1024)
	fmt.Printf("Sys (OS Memory): %0.2f MB (RAM reserved from the OS)\n", float64(m.Sys)/1024/1024)
	fmt.Printf("GC Cycles Run:   %d times (Garbage collector)\n", m.NumGC)
}

func main() {
	// create a cache
	cache := internal.NewCache()
	// // add key-value pairs to it
	// cache.SET("age", 24, 50*time.Second)
	// // Retrieve values
	// key := "age"
	// val, ok := cache.GET(key)
	// if ok == true {
	// 	fmt.Printf("key: %s \tval: %v\n", key, val)
	// }
	// key = "name"
	// val, ok = cache.GET(key)
	// if ok == true {
	// 	fmt.Printf("key: %s \tval: %v\n", key, val)
	// }

	printMemoryStats("Before testing...")
	go cache.Cleanup(50 * time.Millisecond)

	fmt.Println("Running 5 workers x 6000 iterations (180,000 total cache operations)...")

	var wg sync.WaitGroup
	workersCount := 5
	iterationsPerWorker := 6000
	// goroutine to write to the WAL log file
	go cache.WriteToFile()

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
				// We pick random key numbers from 0 to 20 to ensure keys collide across workers.
				// For lookups, we occasionally look up keys 0 to 30 to simulate keys that don't exist.
				randomSetKey := fmt.Sprintf("account_%d", r.Intn(20))
				randomGetKey := fmt.Sprintf("account_%d", r.Intn(30)) 
				randomDelKey := fmt.Sprintf("account_%d", r.Intn(15))

				// SET a dynamic key with varying values and 40ms TTL
				randomVal := r.Float64() * 5000
				cache.SET(randomSetKey, randomVal, 40*time.Millisecond)

				// GET a key (may exist, may not exist, or might have expired)
				cache.GET(randomGetKey)

				// SET value to a heavily used key
				cache.SET("master_ledger", r.Intn(99999), 100*time.Millisecond)

				// read the key
				cache.GET(randomGetKey)

				// Delete a random key
				cache.DEL(randomDelKey)

				// again set a vlaue to random key
				cache.GET(randomSetKey)

				// printing memory stats midway for each worker
				if (workerID == 1 || workerID == 3) && j == iterationsPerWorker/2 {
					printMemoryStats("Mid-operation memory check: ")
				}
			}
			fmt.Printf("Worker %d successfully processed %d iterations\n", workerID, iterationsPerWorker)
		}(i)
	}
	// block the main execution thread until all 5 workers finish their tasks
	startTime := time.Now()
	wg.Wait()
	duration := time.Since(startTime)
	fmt.Printf("\nAll concurrent operations completed cleanly in %v!\n", duration)

	// final verification of the background cleanup process
	printMemoryStats("All operations completed...")
	fmt.Println("\nVerifying final background state...")
	
	// adding fresh keys
	cache.SET("final_check_1", "alive", 100*time.Millisecond)
	cache.SET("final_check_2", "expired_soon", 10*time.Millisecond)

	// Let the cleanup and let keys naturally pass their TTL
	time.Sleep(200 * time.Millisecond)
	_, ok1 := cache.GET("final_check_1")
	_, ok2 := cache.GET("final_check_2")

	fmt.Printf("Key 1 active check (Expected false): %v\n", ok1)
	fmt.Printf("Key 2 active check (Expected false): %v\n", ok2)
}