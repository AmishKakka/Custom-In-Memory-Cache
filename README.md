# Custom In-Memory Cache

A Go-based in-memory cache that supports TTL-based expiration, background cleanup, and concurrent access.

## What it does?

- Thread-safe cache access using Go's `sync.RWMutex`
- `Set(key, value, ttl)` for storing values with a time-to-live
- `Get(key)` for retrieving values only if the key is still active
- `Delete(key)` for removing entries manually
- Background `Cleanup(interval)` loop that removes expired keys
- Simple concurrency test using 5 workers and 6000 iterations each
- WAL (Write-Ahead Logging) every Set and Delete update to keys into a append-only file.

From the project root:

```bash
go run ./cmd/main.go
```

## Output

**Without WAL**
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

**With WAL(Write Ahead Logging)**
```
text
Memory Stats [Before testing...]:
Heap Alloc:      0.15 MB (Active data in RAM)
Total Alloc:     0.15 MB (Cumulative lifetime memory churn)
Sys (OS Memory): 7.77 MB (RAM reserved from the OS)
GC Cycles Run:   0 times (Garbage collector)
Running 5 workers x 6000 iterations (180,000 total cache operations)...

Memory Stats [Mid-operation memory check: ]:
Heap Alloc:      0.37 MB (Active data in RAM)
Total Alloc:     3.57 MB (Cumulative lifetime memory churn)
Sys (OS Memory): 12.85 MB (RAM reserved from the OS)
GC Cycles Run:   1 times (Garbage collector)

Memory Stats [Mid-operation memory check: ]:
Heap Alloc:      0.37 MB (Active data in RAM)
Total Alloc:     3.57 MB (Cumulative lifetime memory churn)
Sys (OS Memory): 12.85 MB (RAM reserved from the OS)
GC Cycles Run:   1 times (Garbage collector)
Worker 5 successfully processed 6000 iterations
Worker 2 successfully processed 6000 iterations
Worker 3 successfully processed 6000 iterations
Worker 4 successfully processed 6000 iterations
Worker 1 successfully processed 6000 iterations

All concurrent operations completed cleanly in 7m9.974588708s!

Memory Stats [All operations completed...]:
Heap Alloc:      0.53 MB (Active data in RAM)
Total Alloc:     6.96 MB (Cumulative lifetime memory churn)
Sys (OS Memory): 12.92 MB (RAM reserved from the OS)
GC Cycles Run:   3 times (Garbage collector)
```

The significant change in time taken for the whole simulation is due to writing to a file (WAL) and making sure it is persisted to disk, but it does guarantee **durability**.

## Notes

- ~~Will work on WAL (Write-Ahead-Log), that writes to a file concurrently.~~
- This cache is in-memory only; it does not persist data to disk. Will do this later.
- Expired entries are cleaned up by the background goroutine and are treated as non-existent by `Get`.
- This designed to demonstrate both correctness and memory behavior under concurrent access.
