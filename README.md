# Text Analyzer

SIS Checkpoint 1 — Golang Application Development.

## Overview

A synchronous Go application that creates and processes text-analysis tasks. Each task has an ID, input text, status, and result. Tasks and results are stored in memory.

## Requirements

Go 1.27.

## Run

From the project root:

```sh
go run ./cmd/demo
```

The demo processes two texts and shows error handling for invalid input, a missing task, and repeated processing.

## Checks

```sh
go test ./...
go vet ./...
gofmt -l .
go build ./...
```

## Structure

- `cmd/demo/main.go` — demonstration.
- `task/task.go` — task model, text analyzer, and in-memory service.
- `task/task_test.go` — unit tests.

## Input and results

Input must be valid UTF-8 and contain at least one non-whitespace character. The original text is preserved.

Results contain:

- Words: whitespace-separated fields.
- Characters: Unicode code points, including spaces and punctuation.
- Letters: Unicode letters.
- Digits: Unicode digits.

For `Hello Go world`, the result is 3 words, 14 characters, 12 letters, and 0 digits.

## Lifecycle

`pending -> processing -> completed`

Create validates input and stores a pending task. Process accepts only pending tasks, changes the status to processing, calculates the result, and saves the completed task. Processing an existing processing or completed task returns an error.

Invalid input returns ErrInvalidInput. Unknown IDs return ErrNotFound. Invalid processing operations return ErrInvalidState.

## Storage

Service uses a map keyed by integer IDs. IDs are unique within a service instance. Get returns a task copy. Initialize the service with NewService.

Processing is synchronous. The service is not concurrency-safe. Data is lost when the application exits.
