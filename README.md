# Custom In-Memory Cache

A Go-based in-memory cache that supports TTL-based expiration, background cleanup, and concurrent access.

## What it does?

- Thread-safe cache access using Go's `sync.RWMutex`
- `Set(key, value, ttl)` for storing values with a time-to-live
- `Get(key)` for retrieving values only if the key is still active
- `Delete(key)` for removing entries manually
- Background `Cleanup(interval)` loop that removes expired keys
- Simple concurrency test using 5 workers and 6000 iterations each

From the project root:

```bash
go run ./cmd/main.go
```

## Output

```text
Memory Stats [Before testing...]:
Heap Alloc:      0.15 MB (Active data in RAM)
Total Alloc:     0.15 MB (Cumulative lifetime memory churn)
Sys (OS Memory): 7.52 MB (RAM reserved from the OS)
GC Cycles Run:   0 times (Garbage collector)

Running 5 workers x 6000 iterations (180,000 total cache operations)...

Memory Stats [Mid-operation memory check: ]:
Heap Alloc:      2.42 MB (Active data in RAM)
Total Alloc:     2.42 MB (Cumulative lifetime memory churn)
Sys (OS Memory): 12.27 MB (RAM reserved from the OS)
GC Cycles Run:   0 times (Garbage collector)

Worker 2 successfully processed 6000 iterations
Worker 3 successfully processed 6000 iterations
Worker 4 successfully processed 6000 iterations
Worker 1 successfully processed 6000 iterations
Worker 5 successfully processed 6000 iterations

All concurrent operations completed cleanly in 41.255916ms!

Memory Stats [All operations completed...]:
Heap Alloc:      1.32 MB (Active data in RAM)
Total Alloc:     4.78 MB (Cumulative lifetime memory churn)
Sys (OS Memory): 13.08 MB (RAM reserved from the OS)
GC Cycles Run:   1 times (Garbage collector)
```

## Notes

- Will work on WAL (Write-Ahead-Log), that writes to a file concurrently.
- This cache is in-memory only; it does not persist data to disk. Will do this later.
- Expired entries are cleaned up by the background goroutine and are treated as non-existent by `Get`.
- This designed to demonstrate both correctness and memory behavior under concurrent access.
