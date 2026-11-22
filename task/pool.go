package task

import (
	"errors"
	"fmt"
	"sync"
)

var (
	ErrQueueFull  = errors.New("task queue is full")
	ErrPoolClosed = errors.New("worker pool is closed")
)

// Pool processes tasks from a bounded queue using a fixed number of workers.
// The service lock guards closed, queue sends and closure, and task transitions.
type Pool struct {
	service *Service
	queue   chan int
	workers sync.WaitGroup
	closed  bool
}

// NewPool starts workerCount workers with queueCapacity waiting slots.
func NewPool(service *Service, workerCount, queueCapacity int) (*Pool, error) {
	if service == nil || service.tasks == nil {
		return nil, errors.New("worker pool requires a service from NewService")
	}
	if workerCount < 2 {
		return nil, errors.New("worker pool requires at least two workers")
	}
	if queueCapacity < 1 {
		return nil, errors.New("task queue capacity must be positive")
	}

	p := &Pool{service: service, queue: make(chan int, queueCapacity)}
	p.workers.Add(workerCount)
	for i := 0; i < workerCount; i++ {
		go p.work()
	}
	return p, nil
}

// Submit accepts a pending task without waiting for queue space or its result.
func (p *Pool) Submit(id int) error {
	s := p.service
	s.mu.Lock()
	defer s.mu.Unlock()
	if p.closed {
		return ErrPoolClosed
	}
	t, err := s.pendingTask(id)
	if err != nil {
		return err
	}

	select {
	case p.queue <- id:
		// A worker can receive the ID now, but it needs mu to read this task.
		t.Status = Queued
		s.tasks[id] = t
		return nil
	default:
		return fmt.Errorf("%w: id=%d", ErrQueueFull, id)
	}
}

func (p *Pool) work() {
	defer p.workers.Done()
	s := p.service
	for id := range p.queue {
		s.mu.Lock()
		t := s.tasks[id]
		t.Status = Processing
		s.tasks[id] = t
		s.mu.Unlock()

		s.complete(t)
	}
}

// Shutdown closes submission, drains accepted jobs, and waits for all workers.
// Repeated or concurrent calls wait for the same workers to exit.
func (p *Pool) Shutdown() {
	p.service.mu.Lock()
	if !p.closed {
		p.closed = true
		close(p.queue)
	}
	p.service.mu.Unlock()
	p.workers.Wait()
}
