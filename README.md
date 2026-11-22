# Text Analyzer

SIS Checkpoint 2 · Golang Application Development

Text Analyzer processes UTF-8 text and counts whitespace-separated words, Unicode code points (including spaces and punctuation), Unicode letters, and Unicode digits. For example, `Сәлем Go 2026` yields 3 words, 13 characters, 7 letters, and 4 digits. The original text, task status, and result are stored in memory under a unique ID.

## Run and verify

Use Go 1.27. From the project root, run:

```sh
go run ./cmd/demo
gofmt -l .
go vet ./...
go test ./...
go test -race ./...
go build ./...
```

The demo creates and submits three texts, waits for the pool to finish, and prints the stored results and sample errors. The race test requires a platform supported by Go's race detector and, on Linux, a C compiler.

## Create and submit tasks

1. Create a service with `task.NewService()`, then start a worker pool with `task.NewPool(service, 2, 3)`.
2. Call `service.Create(text)` to validate and store a task. Empty, whitespace-only, or invalid UTF-8 input returns `ErrInvalidInput`.
3. Call `pool.Submit(created.ID)` to enqueue a pending task. A successful call returns as soon as the task is accepted.
4. Read a snapshot with `service.Get(id)` during or after processing. Unknown IDs return `ErrNotFound`.
5. Call `pool.Shutdown()` to finish the accepted work and wait for the workers to exit. Completed results remain available through `Get`.

## Workers and queue

The demo uses **2 workers** and a buffered channel with **3 waiting slots**. Both values are configurable at startup. The worker count must be at least 2 and the queue capacity must be positive. Each worker repeatedly receives task IDs from the same queue and analyzes the corresponding text. Tasks can finish in any order.

A full queue causes `Submit` to return `ErrQueueFull` immediately. The task stays pending and can be submitted again once a slot is available.

The asynchronous lifecycle is `pending → queued → processing → completed`. Submitting a queued, processing, or completed task returns `ErrInvalidState`. The Checkpoint 1 method `service.Process(id)` is also available for synchronous callers and only accepts pending tasks. Both entry points claim a task under the same lock.

## Synchronization and shutdown

The service's `sync.RWMutex` protects the map, ID allocation, and task transitions. It also protects the pool's closed flag and serializes queue sends with channel closure. A worker acquires that same lock before marking a dequeued task as processing, so it sees the queued state even if it receives the ID immediately. Workers analyze their local copies outside the lock, then save the result and completed status together. `Get` takes a read lock and returns a value copy.

`Shutdown` closes the queue under the write lock, then releases the lock before waiting on a `sync.WaitGroup`. Workers drain the closed channel and exit when it is empty. Each worker signals its exit to the wait group. Repeated or concurrent shutdown calls are safe, and subsequent submissions return `ErrPoolClosed`.

## Project layout

- `cmd/demo/main.go` runs the concurrent text analysis example.
- `task/task.go` defines the task model, text analyzer, and in-memory service.
- `task/pool.go` manages the workers, queue, and shutdown.
- `task/task_test.go` and `task/pool_test.go` verify text analysis and concurrent behavior.
