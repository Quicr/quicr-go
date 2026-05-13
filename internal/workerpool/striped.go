// SPDX-FileCopyrightText: Copyright (c) 2025 Cisco Systems
// SPDX-License-Identifier: BSD-2-Clause

package workerpool

import (
	"runtime"
	"sync"
	"sync/atomic"
)

// StripedPool distributes tasks across multiple independent pools to reduce
// channel contention under high parallelism. Each stripe has its own channel
// and workers, eliminating the single-channel bottleneck.
type StripedPool struct {
	stripes  []*Pool
	counter  atomic.Uint64
	closed   atomic.Bool
	closeMu  sync.Mutex
}

// DefaultStripes returns a reasonable default number of stripes based on CPU count.
func DefaultStripes() int {
	n := runtime.NumCPU()
	if n < 4 {
		return 4
	}
	if n > 16 {
		return 16
	}
	return n
}

// NewStriped creates a new striped worker pool.
// If stripes is 0, DefaultStripes() is used.
// If workersPerStripe is 0, 2 workers per stripe is used.
// If queueSizePerStripe is 0, workersPerStripe*128 is used.
func NewStriped(stripes, workersPerStripe, queueSizePerStripe int) *StripedPool {
	if stripes <= 0 {
		stripes = DefaultStripes()
	}
	if workersPerStripe <= 0 {
		workersPerStripe = 2
	}
	if queueSizePerStripe <= 0 {
		queueSizePerStripe = workersPerStripe * 128
	}

	p := &StripedPool{
		stripes: make([]*Pool, stripes),
	}

	for i := 0; i < stripes; i++ {
		p.stripes[i] = New(workersPerStripe, queueSizePerStripe)
	}

	return p
}

// Submit adds a task to one of the pool's stripes using round-robin distribution.
// If the selected stripe's queue is full, it falls back to a new goroutine.
func (p *StripedPool) Submit(task func()) {
	if p.closed.Load() {
		return
	}
	idx := p.counter.Add(1) % uint64(len(p.stripes))
	p.stripes[idx].Submit(task)
}

// SubmitTo adds a task to a specific stripe, useful when ordering matters
// for tasks with the same key. The key is hashed to select a stripe.
func (p *StripedPool) SubmitTo(key uint64, task func()) {
	if p.closed.Load() {
		return
	}
	idx := key % uint64(len(p.stripes))
	p.stripes[idx].Submit(task)
}

// TrySubmit attempts to add a task using round-robin distribution.
// Returns true if queued, false if all attempted stripes are full.
func (p *StripedPool) TrySubmit(task func()) bool {
	if p.closed.Load() {
		return false
	}
	idx := p.counter.Add(1) % uint64(len(p.stripes))
	return p.stripes[idx].TrySubmit(task)
}

// Close shuts down all stripes after their queued tasks complete.
func (p *StripedPool) Close() {
	p.closeMu.Lock()
	defer p.closeMu.Unlock()

	if p.closed.Swap(true) {
		return
	}

	for _, stripe := range p.stripes {
		stripe.Close()
	}
}

// NumStripes returns the number of stripes in the pool.
func (p *StripedPool) NumStripes() int {
	return len(p.stripes)
}

// TotalWorkers returns the total number of workers across all stripes.
func (p *StripedPool) TotalWorkers() int {
	total := 0
	for _, stripe := range p.stripes {
		total += stripe.workers
	}
	return total
}
