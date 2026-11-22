package task

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"
)

type analyzerFunc func(string) Result

func (f analyzerFunc) Analyze(text string) Result {
	return f(text)
}

func receive[T any](t *testing.T, ch <-chan T) T {
	t.Helper()
	select {
	case value := <-ch:
		return value
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for concurrent operation")
		var zero T
		return zero
	}
}

func TestPoolConfiguration(t *testing.T) {
	for _, config := range [][2]int{{0, 2}, {1, 2}, {2, 0}, {2, -1}} {
		if p, err := NewPool(NewService(), config[0], config[1]); err == nil {
			p.Shutdown()
			t.Errorf("NewPool accepted workers=%d, capacity=%d", config[0], config[1])
		}
	}
}

func TestPoolOverlapsWorkAndDrainsQueue(t *testing.T) {
	s := NewService()
	inputs := []string{"Go 12", "Сәлем!", "two words", "☀"}
	want := []Result{{2, 5, 2, 2}, {1, 6, 5, 0}, {2, 9, 8, 0}}
	var created []Task
	for _, input := range inputs {
		item, err := s.Create(input)
		if err != nil {
			t.Fatal(err)
		}
		created = append(created, item)
	}

	started := make(chan string, len(inputs))
	release := make(chan struct{})
	s.analyzer = analyzerFunc(func(text string) Result {
		started <- text
		<-release
		return TextAnalyzer{}.Analyze(text)
	})
	p, err := NewPool(s, 2, 1)
	if err != nil {
		t.Fatal(err)
	}
	var releaseOnce sync.Once
	unblock := func() { releaseOnce.Do(func() { close(release) }) }
	defer p.Shutdown()
	defer unblock()

	for _, item := range created[:2] {
		if err := p.Submit(item.ID); err != nil {
			t.Fatal(err)
		}
		// Workers report entry while blocked on release, so both analyses overlap.
		if input := receive(t, started); input != item.Input {
			t.Fatalf("worker received %q, want %q", input, item.Input)
		}
	}
	for _, item := range created[:2] {
		snapshot, err := s.Get(item.ID)
		if err != nil || snapshot.Status != Processing || snapshot.Result != (Result{}) {
			t.Fatalf("Get(%d) while analyzing = %+v, %v", item.ID, snapshot, err)
		}
		if err := p.Submit(item.ID); !errors.Is(err, ErrInvalidState) {
			t.Fatalf("resubmitting processing task: %v", err)
		}
	}
	if err := p.Submit(created[2].ID); err != nil {
		t.Fatal(err)
	}
	queued, err := s.Get(created[2].ID)
	if err != nil || queued.Status != Queued {
		t.Fatalf("queued task = %+v, %v", queued, err)
	}
	if err := p.Submit(queued.ID); !errors.Is(err, ErrInvalidState) {
		t.Fatalf("resubmitting queued task: %v", err)
	}
	if _, err := s.Process(queued.ID); !errors.Is(err, ErrInvalidState) {
		t.Fatalf("synchronously processing queued task: %v", err)
	}
	if err := p.Submit(created[3].ID); !errors.Is(err, ErrQueueFull) {
		t.Fatalf("full queue: %v", err)
	}
	rejected, err := s.Get(created[3].ID)
	if err != nil || rejected != created[3] {
		t.Fatalf("rejected task changed: %+v, %v", rejected, err)
	}
	if err := p.Submit(0); !errors.Is(err, ErrNotFound) {
		t.Fatalf("submitting missing task: %v", err)
	}

	stopped := make(chan struct{})
	go func() {
		p.Shutdown()
		close(stopped)
	}()
	unblock()
	receive(t, stopped)
	p.Shutdown()
	if _, ok := <-p.queue; ok {
		t.Fatal("queue still contains work after shutdown")
	}
	for i, item := range created[:3] {
		got, err := s.Get(item.ID)
		if err != nil || got.ID != item.ID || got.Input != item.Input || got.Status != Completed || got.Result != want[i] {
			t.Fatalf("Get(%d) = %+v, %v; want completed task with result %+v", item.ID, got, err, want[i])
		}
	}
	if err := p.Submit(rejected.ID); !errors.Is(err, ErrPoolClosed) {
		t.Fatalf("submitting after shutdown: %v", err)
	}
}

func TestPoolConcurrentDuplicateSubmission(t *testing.T) {
	s := NewService()
	item, err := s.Create("Go")
	if err != nil {
		t.Fatal(err)
	}
	p, err := NewPool(s, 2, 2)
	if err != nil {
		t.Fatal(err)
	}
	defer p.Shutdown()

	const attempts = 16
	start := make(chan struct{})
	results := make(chan error, attempts)
	for i := 0; i < attempts; i++ {
		go func() {
			<-start
			results <- p.Submit(item.ID)
		}()
	}
	close(start)
	accepted := 0
	for i := 0; i < attempts; i++ {
		err := receive(t, results)
		if err == nil {
			accepted++
		} else if !errors.Is(err, ErrInvalidState) {
			t.Errorf("unexpected duplicate submission error: %v", err)
		}
	}
	if accepted != 1 {
		t.Fatalf("accepted %d copies of the same task, want 1", accepted)
	}
	p.Shutdown()

	// A new pool for the same service must also reject completed tasks.
	next, err := NewPool(s, 2, 2)
	if err != nil {
		t.Fatal(err)
	}
	defer next.Shutdown()
	if err := next.Submit(item.ID); !errors.Is(err, ErrInvalidState) {
		t.Fatalf("submitting completed task: %v", err)
	}
}

func TestPoolConcurrentStorageAndShutdown(t *testing.T) {
	const count = 32
	s := NewService()
	p, err := NewPool(s, 3, count)
	if err != nil {
		t.Fatal(err)
	}
	defer p.Shutdown()

	type outcome struct {
		created   Task
		submitErr error
	}
	outcomes := make(chan outcome, count)
	firstSubmission := make(chan struct{})
	ready := make(chan struct{})
	for i := 0; i < count; i++ {
		go func() {
			<-ready
			input := fmt.Sprintf("task %d", i)
			item, err := s.Create(input)
			if err != nil {
				t.Errorf("Create(%q): %v", input, err)
				outcomes <- outcome{}
				if i == 0 {
					close(firstSubmission)
				}
				return
			}
			submitErr := p.Submit(item.ID)
			if i == 0 {
				close(firstSubmission)
			}
			if got, err := s.Get(item.ID); err != nil || got.ID != item.ID || got.Input != input {
				t.Errorf("concurrent Get(%d) = %+v, %v", item.ID, got, err)
			}
			outcomes <- outcome{item, submitErr}
		}()
	}
	stopped := make(chan struct{})
	go func() {
		<-firstSubmission
		p.Shutdown()
		close(stopped)
	}()
	close(ready)
	receive(t, stopped)

	seen := make(map[int]bool)
	accepted := 0
	for i := 0; i < count; i++ {
		out := receive(t, outcomes)
		id := out.created.ID
		if id < 1 || id > count || seen[id] {
			t.Errorf("invalid or duplicate ID %d", id)
		}
		seen[id] = true
		got, err := s.Get(id)
		if err != nil {
			t.Error(err)
			continue
		}
		switch {
		case out.submitErr == nil:
			accepted++
			want := TextAnalyzer{}.Analyze(out.created.Input)
			if got.Status != Completed || got.Result != want {
				t.Errorf("accepted task %d = %+v, want completed with %+v", id, got, want)
			}
		case errors.Is(out.submitErr, ErrPoolClosed):
			if got != out.created {
				t.Errorf("rejected task %d changed: %+v", id, got)
			}
		default:
			t.Errorf("unexpected submit error: %v", out.submitErr)
		}
	}
	if accepted == 0 {
		t.Fatal("no submissions were accepted")
	}
}
