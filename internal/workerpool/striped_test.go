// SPDX-FileCopyrightText: Copyright (c) 2025 Cisco Systems
// SPDX-License-Identifier: BSD-2-Clause

package workerpool

import (
	"sync"
	"sync/atomic"
	"testing"
)

func TestNewStriped(t *testing.T) {
	tests := []struct {
		name               string
		stripes            int
		workersPerStripe   int
		queueSizePerStripe int
		wantStripes        int
	}{
		{
			name:        "defaults",
			stripes:     0,
			wantStripes: DefaultStripes(),
		},
		{
			name:        "custom stripes",
			stripes:     8,
			wantStripes: 8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewStriped(tt.stripes, tt.workersPerStripe, tt.queueSizePerStripe)
			defer p.Close()

			if got := p.NumStripes(); got != tt.wantStripes {
				t.Errorf("NumStripes() = %d, want %d", got, tt.wantStripes)
			}
		})
	}
}

func TestStripedPool_Submit(t *testing.T) {
	p := NewStriped(4, 2, 64)
	defer p.Close()

	var count atomic.Int32
	var wg sync.WaitGroup

	n := 1000
	wg.Add(n)
	for i := 0; i < n; i++ {
		p.Submit(func() {
			count.Add(1)
			wg.Done()
		})
	}

	wg.Wait()

	if got := count.Load(); got != int32(n) {
		t.Errorf("executed %d tasks, want %d", got, n)
	}
}

func TestStripedPool_SubmitTo(t *testing.T) {
	p := NewStriped(4, 2, 64)
	defer p.Close()

	// Tasks with the same key should go to the same stripe
	var order []int
	var mu sync.Mutex
	var wg sync.WaitGroup

	n := 100
	wg.Add(n)
	for i := 0; i < n; i++ {
		i := i
		p.SubmitTo(42, func() { // Same key = same stripe = ordered execution
			mu.Lock()
			order = append(order, i)
			mu.Unlock()
			wg.Done()
		})
	}

	wg.Wait()

	// Verify all tasks executed
	if len(order) != n {
		t.Errorf("executed %d tasks, want %d", len(order), n)
	}
}

func TestStripedPool_Close(t *testing.T) {
	p := NewStriped(4, 2, 64)

	var count atomic.Int32
	var wg sync.WaitGroup

	// Submit some tasks
	n := 100
	wg.Add(n)
	for i := 0; i < n; i++ {
		p.Submit(func() {
			count.Add(1)
			wg.Done()
		})
	}

	// Close should wait for all tasks
	p.Close()
	wg.Wait()

	if got := count.Load(); got != int32(n) {
		t.Errorf("executed %d tasks, want %d", got, n)
	}

	// Double close should be safe
	p.Close()
}

func TestStripedPool_SubmitAfterClose(t *testing.T) {
	p := NewStriped(4, 2, 64)
	p.Close()

	// Should not panic
	p.Submit(func() {
		t.Error("task should not execute after close")
	})
}

func BenchmarkStripedPool_Submit(b *testing.B) {
	p := NewStriped(0, 0, 0)
	defer p.Close()

	done := make(chan struct{})
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		p.Submit(func() {
			done <- struct{}{}
		})
		<-done
	}
}

func BenchmarkStripedPool_Submit_Parallel(b *testing.B) {
	p := NewStriped(0, 0, 0)
	defer p.Close()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		done := make(chan struct{})
		for pb.Next() {
			p.Submit(func() {
				done <- struct{}{}
			})
			<-done
		}
	})
}

func BenchmarkStripedPool_SubmitTo_Parallel(b *testing.B) {
	p := NewStriped(0, 0, 0)
	defer p.Close()

	var key atomic.Uint64

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		done := make(chan struct{})
		myKey := key.Add(1)
		for pb.Next() {
			p.SubmitTo(myKey, func() {
				done <- struct{}{}
			})
			<-done
		}
	})
}

func BenchmarkStripedPool_Throughput(b *testing.B) {
	p := NewStriped(0, 0, 0)
	defer p.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.Submit(func() {})
	}
}

func BenchmarkStripedPool_Throughput_Parallel(b *testing.B) {
	p := NewStriped(0, 0, 0)
	defer p.Close()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			p.Submit(func() {})
		}
	})
}

// Compare striped pool vs regular pool under parallel load
func BenchmarkComparison_RegularPool_Parallel(b *testing.B) {
	p := New(DefaultWorkers(), DefaultWorkers()*256)
	defer p.Close()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			p.Submit(func() {})
		}
	})
}

func BenchmarkComparison_StripedPool_Parallel(b *testing.B) {
	p := NewStriped(0, 0, 0)
	defer p.Close()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			p.Submit(func() {})
		}
	})
}
