// SPDX-FileCopyrightText: Copyright (c) 2025 Cisco Systems
// SPDX-License-Identifier: BSD-2-Clause

// Package workerpool provides a simple worker pool for executing callbacks
// without creating a new goroutine for each callback.
package workerpool

import (
	"errors"
	"runtime"
	"sync"
	"sync/atomic"
)

// ErrPoolClosed is returned when submitting to a closed pool.
var ErrPoolClosed = errors.New("workerpool: pool is closed")

// ErrQueueFull is returned when the queue is full and backpressure is enabled.
var ErrQueueFull = errors.New("workerpool: queue is full")

// Pool manages a pool of worker goroutines that process tasks from a shared queue.
// It's optimized for high-frequency callback dispatch.
type Pool struct {
	tasks   chan func()
	wg      sync.WaitGroup
	once    sync.Once
	workers int
	closed  atomic.Bool
}

// DefaultWorkers returns a reasonable default number of workers based on CPU count.
func DefaultWorkers() int {
	n := runtime.NumCPU()
	if n < 4 {
		return 4
	}
	if n > 16 {
		return 16
	}
	return n
}

// New creates a new worker pool with the specified number of workers and queue size.
// If workers is 0, DefaultWorkers() is used.
// If queueSize is 0, a default of workers*256 is used.
func New(workers, queueSize int) *Pool {
	if workers <= 0 {
		workers = DefaultWorkers()
	}
	if queueSize <= 0 {
		queueSize = workers * 256
	}

	p := &Pool{
		tasks:   make(chan func(), queueSize),
		workers: workers,
	}

	p.wg.Add(workers)
	for i := 0; i < workers; i++ {
		go p.worker()
	}

	return p
}

// worker processes tasks from the queue.
func (p *Pool) worker() {
	defer p.wg.Done()
	for task := range p.tasks {
		task()
	}
}

// Submit adds a task to the pool's queue.
// If the queue is full, it falls back to running the task in a new goroutine.
// If the pool is closed, the task is silently dropped.
func (p *Pool) Submit(task func()) {
	if p.closed.Load() {
		return
	}
	select {
	case p.tasks <- task:
		// Task queued successfully
	default:
		// Queue full - fall back to goroutine
		go task()
	}
}

// SubmitWait adds a task to the pool's queue, blocking until space is available.
// Returns ErrPoolClosed if the pool has been closed.
func (p *Pool) SubmitWait(task func()) error {
	if p.closed.Load() {
		return ErrPoolClosed
	}
	select {
	case p.tasks <- task:
		return nil
	default:
		// Channel might be closed, check again
		if p.closed.Load() {
			return ErrPoolClosed
		}
		// Block until space is available
		p.tasks <- task
		return nil
	}
}

// SubmitWithBackpressure adds a task to the pool's queue.
// Returns ErrQueueFull if the queue is full (task is NOT executed).
// Returns ErrPoolClosed if the pool has been closed.
func (p *Pool) SubmitWithBackpressure(task func()) error {
	if p.closed.Load() {
		return ErrPoolClosed
	}
	select {
	case p.tasks <- task:
		return nil
	default:
		return ErrQueueFull
	}
}

// TrySubmit attempts to add a task to the pool's queue.
// Returns true if queued, false if the queue is full (task is NOT executed).
func (p *Pool) TrySubmit(task func()) bool {
	if p.closed.Load() {
		return false
	}
	select {
	case p.tasks <- task:
		return true
	default:
		return false
	}
}

// Close shuts down the pool after all queued tasks complete.
// New submissions after Close are silently dropped.
func (p *Pool) Close() {
	p.once.Do(func() {
		p.closed.Store(true)
		close(p.tasks)
		p.wg.Wait()
	})
}

// IsClosed returns true if the pool has been closed.
func (p *Pool) IsClosed() bool {
	return p.closed.Load()
}

// Workers returns the number of workers in the pool.
func (p *Pool) Workers() int {
	return p.workers
}
