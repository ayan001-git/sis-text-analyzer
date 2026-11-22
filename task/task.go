// Package task implements in-memory text analysis and concurrent task processing.
package task

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"
)

type Status string

const (
	Pending    Status = "pending"
	Queued     Status = "queued"
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
// Analyze must be deterministic and safe to call concurrently.
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

// Service owns all stored tasks and protects them with mu.
// Returned Task values are snapshots, so callers cannot mutate stored state.
type Service struct {
	mu       sync.RWMutex
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
	s.mu.Lock()
	defer s.mu.Unlock()
	t := Task{ID: s.nextID, Input: text, Status: Pending}
	s.tasks[t.ID] = t
	s.nextID++
	return t, nil
}

func (s *Service) Get(id int) (Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.get(id)
}

// get requires the caller to hold mu for reading or writing.
func (s *Service) get(id int) (Task, error) {
	t, ok := s.tasks[id]
	if !ok {
		return Task{}, fmt.Errorf("%w: id=%d", ErrNotFound, id)
	}
	return t, nil
}

// pendingTask requires the caller to hold mu when checking and claiming a task.
func (s *Service) pendingTask(id int) (Task, error) {
	t, err := s.get(id)
	if err != nil {
		return Task{}, err
	}
	if t.Status != Pending {
		return Task{}, fmt.Errorf("%w: id=%d status=%s", ErrInvalidState, id, t.Status)
	}
	return t, nil
}

// Process analyzes a pending task synchronously, as in Checkpoint 1.
func (s *Service) Process(id int) (Task, error) {
	s.mu.Lock()
	t, err := s.pendingTask(id)
	if err != nil {
		s.mu.Unlock()
		return Task{}, err
	}
	t.Status = Processing
	s.tasks[id] = t
	s.mu.Unlock()
	return s.complete(t), nil
}

// complete runs analysis on a local copy before publishing the result and status together.
func (s *Service) complete(t Task) Task {
	t.Result = s.analyzer.Analyze(t.Input)
	t.Status = Completed
	s.mu.Lock()
	s.tasks[t.ID] = t
	s.mu.Unlock()
	return t
}
