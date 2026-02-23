package worker

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/solo-ai/solo/pkg/types"
)

// Worker executes tasks. Lightweight alternative to Ray actors.
type Worker struct {
	id        string
	executor  types.TaskExecutor
	running   atomic.Bool
	done      chan struct{}
	waitGroup sync.WaitGroup
}

// New creates a new worker.
func New(id string, executor types.TaskExecutor) *Worker {
	return &Worker{
		id:       id,
		executor: executor,
		done:     make(chan struct{}),
	}
}

// ID returns the worker identifier.
func (w *Worker) ID() string {
	return w.id
}

// Run executes a task and returns the result.
func (w *Worker) Run(ctx context.Context, task *types.Task) (*types.Task, error) {
	w.running.Store(true)
	defer w.running.Store(false)

	return w.executor.Execute(ctx, task)
}

// IsRunning reports whether the worker is currently executing.
func (w *Worker) IsRunning() bool {
	return w.running.Load()
}

// Pool manages a pool of workers for parallel execution.
type Pool struct {
	workers   []*Worker
	sem       chan struct{}
	executor  types.TaskExecutor
}

// NewPool creates a worker pool with the given size.
func NewPool(size int, executor types.TaskExecutor) *Pool {
	sem := make(chan struct{}, size)
	workers := make([]*Worker, size)
	for i := 0; i < size; i++ {
		workers[i] = New(string(rune('a'+i)), executor)
	}
	return &Pool{
		workers:  workers,
		sem:      sem,
		executor: executor,
	}
}

// Submit runs a task when a worker is available.
func (p *Pool) Submit(ctx context.Context, task *types.Task) (*types.Task, error) {
	select {
	case p.sem <- struct{}{}:
		defer func() { <-p.sem }()
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	w := New(task.ID+"-worker", p.executor)
	return w.Run(ctx, task)
}
