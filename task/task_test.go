package task

import (
	"errors"
	"testing"
)

func TestCreateAndGet(t *testing.T) {
	s := NewService()
	a, err := s.Create("Hello Go world")
	if err != nil {
		t.Fatal(err)
	}
	if a.ID != 1 || a.Input != "Hello Go world" || a.Status != Pending || a.Result != (Result{}) {
		t.Fatalf("unexpected created task: %+v", a)
	}
	got, err := s.Get(a.ID)
	if err != nil || got != a {
		t.Fatalf("Get = %+v, %v; want %+v", got, err, a)
	}
	a.Input = "changed"
	a.Status = Completed
	stored, _ := s.Get(1)
	if stored.Input != "Hello Go world" || stored.Status != Pending {
		t.Fatal("caller changed storage")
	}
}

func TestInvalidInput(t *testing.T) {
	for _, input := range []string{"", " ", "\n\t", "\u2003", string([]byte{0xff})} {
		t.Run(input, func(t *testing.T) {
			s := NewService()
			if _, err := s.Create(input); !errors.Is(err, ErrInvalidInput) {
				t.Fatalf("got %v", err)
			}
			if len(s.tasks) != 0 {
				t.Fatal("invalid task was stored")
			}
			valid, err := s.Create("valid")
			if err != nil || valid.ID != 1 {
				t.Fatal("rejected input consumed an ID")
			}
		})
	}
}

func TestMissingTask(t *testing.T) {
	s := NewService()
	for _, id := range []int{-1, 0, 999} {
		if _, err := s.Get(id); !errors.Is(err, ErrNotFound) {
			t.Fatalf("Get(%d): %v", id, err)
		}
		if _, err := s.Process(id); !errors.Is(err, ErrNotFound) {
			t.Fatalf("Process(%d): %v", id, err)
		}
	}
}

func TestProcessResults(t *testing.T) {
	cases := []struct {
		name, input string
		want        Result
	}{
		{"ASCII", "Hello Go world", Result{3, 14, 12, 0}},
		{"Unicode", "Сәлем Go 2026", Result{3, 13, 7, 4}},
		{"Whitespace", "  one\ttwo\n", Result{2, 10, 6, 0}},
		{"Punctuation", "Go, go!", Result{2, 7, 4, 0}},
		{"Emoji", "Hi 🌍", Result{2, 4, 2, 0}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := NewService()
			created, err := s.Create(tc.input)
			if err != nil {
				t.Fatal(err)
			}
			got, err := s.Process(created.ID)
			if err != nil {
				t.Fatal(err)
			}
			if got.Result != tc.want || got.Status != Completed || got.Input != tc.input || got.ID != created.ID {
				t.Fatalf("got %+v; want result %+v and completed", got, tc.want)
			}
			stored, err := s.Get(created.ID)
			if err != nil || stored != got {
				t.Fatalf("result not stored: %+v %v", stored, err)
			}
			if _, err := s.Process(created.ID); !errors.Is(err, ErrInvalidState) {
				t.Fatalf("repeat: %v", err)
			}
			after, _ := s.Get(created.ID)
			if after != got {
				t.Fatal("reprocessing changed task")
			}
		})
	}
}

func TestIndependentTasks(t *testing.T) {
	s := NewService()
	a, _ := s.Create("one")
	b, _ := s.Create("two words")
	if a.ID == b.ID {
		t.Fatal("duplicate IDs")
	}
	doneA, err := s.Process(a.ID)
	if err != nil {
		t.Fatal(err)
	}
	pendingB, _ := s.Get(b.ID)
	if pendingB != b {
		t.Fatal("processing A changed B")
	}
	doneB, err := s.Process(b.ID)
	if err != nil {
		t.Fatal(err)
	}
	if doneA.Result != (Result{1, 3, 3, 0}) || doneB.Result != (Result{2, 9, 8, 0}) {
		t.Fatal("mixed results")
	}
	storedA, _ := s.Get(a.ID)
	if storedA != doneA {
		t.Fatal("processing B changed A")
	}
}

// A test double checks that Processing is actually stored during calculation.
type observingAnalyzer struct {
	t       *testing.T
	service *Service
	id      int
}

func (a observingAnalyzer) Analyze(text string) Result {
	stored, err := a.service.Get(a.id)
	if err != nil || stored.Status != Processing {
		a.t.Fatalf("during analysis: %+v %v", stored, err)
	}
	if _, err := a.service.Process(a.id); !errors.Is(err, ErrInvalidState) {
		a.t.Fatalf("processing state accepted: %v", err)
	}
	return TextAnalyzer{}.Analyze(text)
}
func TestProcessingState(t *testing.T) {
	s := NewService()
	created, _ := s.Create("Go")
	s.analyzer = observingAnalyzer{t, s, created.ID}
	got, err := s.Process(created.ID)
	if err != nil || got.Status != Completed {
		t.Fatalf("got %+v, %v", got, err)
	}
}
