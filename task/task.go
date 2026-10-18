// Package task implements synchronous in-memory text analysis.
package task

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

type Status string

const (
	Pending    Status = "pending"
	Processing Status = "processing"
	Completed  Status = "completed"
)

var (
	ErrInvalidInput = errors.New("text must be valid UTF-8 and contain non-whitespace characters")
	ErrNotFound     = errors.New("task not found")
	ErrInvalidState = errors.New("task cannot be processed in its current state")
)

// Result counts Unicode code points, not bytes or visual grapheme clusters.
// Words are whitespace-separated fields; punctuation stays part of a word.
type Result struct {
	Words      int
	Characters int
	Letters    int
	Digits     int
}

type Task struct {
	ID     int
	Input  string
	Status Status
	Result Result
}

// Analyzer is the processing dependency; another analyzer can be substituted.
// Analyze must be deterministic and must not call back into Service.
type Analyzer interface {
	Analyze(text string) Result
}

type TextAnalyzer struct{}

// A value receiver is sufficient: TextAnalyzer has no mutable fields.
func (TextAnalyzer) Analyze(text string) Result {
	result := Result{Words: len(strings.Fields(text)), Characters: utf8.RuneCountInString(text)}
	for _, r := range text {
		if unicode.IsLetter(r) {
			result.Letters++
		}
		if unicode.IsDigit(r) {
			result.Digits++
		}
	}
	return result
}

// Service owns all stored tasks. It is synchronous and not concurrency-safe.
// Returned Task values are snapshots, so callers cannot mutate stored state.
type Service struct {
	tasks    map[int]Task
	nextID   int
	analyzer Analyzer
}

// NewService supplies a working default and allocates the storage map.
func NewService() *Service {
	return &Service{tasks: make(map[int]Task), nextID: 1, analyzer: TextAnalyzer{}}
}

func (s *Service) Create(text string) (Task, error) {
	if !utf8.ValidString(text) || strings.TrimSpace(text) == "" {
		return Task{}, ErrInvalidInput
	}
	t := Task{ID: s.nextID, Input: text, Status: Pending}
	s.tasks[t.ID] = t
	s.nextID++
	return t, nil
}

func (s *Service) Get(id int) (Task, error) {
	t, ok := s.tasks[id]
	if !ok {
		return Task{}, fmt.Errorf("%w: id=%d", ErrNotFound, id)
	}
	return t, nil
}

// Process stores Processing before invoking the analyzer, and Completed only
// after its result exists. It rejects reprocessing rather than overwriting results.
func (s *Service) Process(id int) (Task, error) {
	t, err := s.Get(id)
	if err != nil {
		return Task{}, err
	}
	if t.Status != Pending {
		return Task{}, fmt.Errorf("%w: id=%d status=%s", ErrInvalidState, id, t.Status)
	}
	t.Status = Processing
	s.tasks[id] = t
	t.Result = s.analyzer.Analyze(t.Input)
	t.Status = Completed
	s.tasks[id] = t
	return t, nil
}
