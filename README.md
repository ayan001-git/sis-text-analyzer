# SIS Checkpoint 1 — Text Analyzer

Author: Tursynmamet Ayan. Individual project for Golang Application Development.

## What it does

A task represents analysis of one text. The service validates input, assigns an integer ID, stores the task in a map, calculates text statistics, and stores the result. No external dependencies, database, HTTP, goroutines, or persistence.

Requires **Go 1.27.0 or later in the Go 1.27 series**, as specified by the assignment.

## Run

From the folder containing go.mod:

```sh
go version
go run ./cmd/demo
go test -v ./...
go vet ./...
gofmt -l .
go build ./...
```

`gofmt -l .` and `go vet ./...` should print nothing. Tests should pass.

## Structure

- `task/task.go`: domain types, errors, Analyzer interface, text analysis, in-memory Service.
- `task/task_test.go`: creation, validation, lookup, results, states, invalid operations, independence, snapshot isolation.
- `cmd/demo/main.go`: runnable demonstration including expected error cases.
- `docs/DEFENSE_RU_EN.md`: explanation, defense questions and submission steps.
- `docs/REQUIREMENTS.md`: specification-to-code mapping and sources.
- `docs/VALIDATION.txt`: actual environment verification results.

## Rules and lifecycle

Input must be valid UTF-8 and contain a non-whitespace character. Original input is preserved.

`pending -> processing -> completed`

Create stores pending. Process accepts only pending, stores processing before analysis, then stores result and completed. Processing a completed or processing task returns ErrInvalidState. There is no public state setter. Unknown IDs return ErrNotFound; invalid text returns ErrInvalidInput. errors.Is supports checking wrapped errors.

Processing is synchronous; after Process returns, the task is completed. The intermediate processing state is verified with an observing analyzer in a unit test. The built-in analyzer is deterministic and has no expected processing failure for validated input, so a failed state is not needed.

## Result definitions

- Words: `len(strings.Fields(text))`; whitespace separates words, punctuation does not.
- Characters: Unicode code points (runes), including spaces and punctuation, not UTF-8 bytes or grapheme clusters.
- Letters: runes for which unicode.IsLetter is true.
- Digits: runes for which unicode.IsDigit is true.

Example: `Hello Go world` gives words=3, characters=14 (5+1+2+1+5), letters=12 (5+2+5), digits=0.

## Design

Service owns a map[int]Task and a monotonically increasing ID counter scoped to one service instance. IDs restart at 1 on restart. NewService allocates the map and supplies TextAnalyzer. Use the constructor; zero-value Service is not supported.

Methods modifying service state use pointer receivers. TextAnalyzer uses a value receiver because it has no mutable state. Analyzer is a meaningful processing abstraction: production uses TextAnalyzer, while tests substitute an observer to check the actual processing state.

Get returns a Task value snapshot. All Task fields are values (including immutable strings and a value Result), so callers cannot mutate stored tasks. Changing a retrieved Task does not change the map.

This implementation is deliberately single-threaded and not concurrency-safe. All data is lost on program exit. Concurrency and HTTP belong to later checkpoints.

## Sources

Specification: https://github.com/salikhP/GOLANG-FALL-2026-KBTU/blob/0d60b05cf2e270f93203fa2b362dd5b0c0305e8f/sis/SIS_CP1_SPECIFICATION.md

Student guidelines: https://github.com/salikhP/GOLANG-FALL-2026-KBTU/blob/0d60b05cf2e270f93203fa2b362dd5b0c0305e8f/sis/SIS%20Student%20Guidelines.pdf

See docs/REQUIREMENTS.md for the full mapping. No score is guaranteed; defense is required.
