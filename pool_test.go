// SPDX-FileCopyrightText: Copyright (c) 2025 Cisco Systems
// SPDX-License-Identifier: BSD-2-Clause

package qgo

import (
	"testing"
)

func TestObjectPool_Get(t *testing.T) {
	pool := DefaultObjectPool

	obj := pool.Get()
	if obj == nil {
		t.Fatal("Get() returned nil")
	}

	// Set some values
	obj.Object.Headers.GroupID = 123
	obj.Object.Headers.ObjectID = 456
	obj.Object.Data = []byte("test data")

	// Release back to pool
	obj.Release()

	// Get again - should be zeroed
	obj2 := pool.Get()
	if obj2.Object.Headers.GroupID != 0 {
		t.Error("GroupID not zeroed after release")
	}
	if obj2.Object.Headers.ObjectID != 0 {
		t.Error("ObjectID not zeroed after release")
	}
	if obj2.Object.Data != nil {
		t.Error("Data not zeroed after release")
	}
	obj2.Release()
}

func TestObjectPool_GetWithData(t *testing.T) {
	pool := DefaultObjectPool

	obj := pool.GetWithData(1024)
	if obj == nil {
		t.Fatal("GetWithData() returned nil")
	}

	if obj.Object.Data == nil {
		t.Fatal("Data is nil")
	}

	if len(obj.Object.Data) != 1024 {
		t.Errorf("Data length = %d, want 1024", len(obj.Object.Data))
	}

	// Fill with data
	for i := range obj.Object.Data {
		obj.Object.Data[i] = byte(i % 256)
	}

	obj.Release()
}

func TestObjectPool_LargeData(t *testing.T) {
	pool := DefaultObjectPool

	// Request larger than maxPooledDataSize
	obj := pool.GetWithData(100000)
	if obj == nil {
		t.Fatal("GetWithData() returned nil")
	}

	if len(obj.Object.Data) != 100000 {
		t.Errorf("Data length = %d, want 100000", len(obj.Object.Data))
	}

	obj.Release()
}

func BenchmarkObjectPool_GetRelease(b *testing.B) {
	pool := DefaultObjectPool

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		obj := pool.Get()
		obj.Release()
	}
}

func BenchmarkObjectPool_GetWithDataRelease(b *testing.B) {
	pool := DefaultObjectPool

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		obj := pool.GetWithData(1024)
		obj.Release()
	}
}

func BenchmarkObjectPool_GetRelease_Parallel(b *testing.B) {
	pool := DefaultObjectPool

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			obj := pool.Get()
			obj.Release()
		}
	})
}

func BenchmarkObjectPool_NoPool(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		obj := &Object{
			Data: make([]byte, 1024),
		}
		_ = obj
	}
}

func BenchmarkObjectPool_WithPool(b *testing.B) {
	pool := DefaultObjectPool

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		obj := pool.GetWithData(1024)
		obj.Release()
	}
}
