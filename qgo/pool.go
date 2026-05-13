// SPDX-FileCopyrightText: Copyright (c) 2025 Cisco Systems
// SPDX-License-Identifier: BSD-2-Clause

package qgo

import (
	"sync"
)

// Object pools for reducing GC pressure in high-throughput scenarios.
// These are used internally when pooling is enabled.

const (
	// defaultDataPoolSize is the default capacity for pooled byte slices.
	defaultDataPoolSize = 4096

	// maxPooledDataSize is the maximum size for pooled byte slices.
	// Larger allocations bypass the pool.
	maxPooledDataSize = 65536
)

// objectPool pools Object structs to reduce allocations.
var objectPool = sync.Pool{
	New: func() any {
		return &PooledObject{
			Object: Object{},
		}
	},
}

// dataPool pools byte slices for object data.
var dataPool = sync.Pool{
	New: func() any {
		b := make([]byte, 0, defaultDataPoolSize)
		return &b
	},
}

// PooledObject wraps an Object with pool management.
// Call Release() when done to return resources to the pool.
type PooledObject struct {
	Object
	pooledData *[]byte
}

// Release returns the object and its data buffer to their respective pools.
// After calling Release, the PooledObject must not be used.
func (p *PooledObject) Release() {
	if p.pooledData != nil {
		// Reset slice but keep capacity
		*p.pooledData = (*p.pooledData)[:0]
		dataPool.Put(p.pooledData)
		p.pooledData = nil
	}
	p.Object = Object{}
	objectPool.Put(p)
}

// acquirePooledObject gets an object from the pool.
func acquirePooledObject() *PooledObject {
	return objectPool.Get().(*PooledObject)
}

// acquireDataBuffer gets a byte slice from the pool with at least the given capacity.
func acquireDataBuffer(size int) *[]byte {
	if size > maxPooledDataSize {
		// Too large for pool, allocate directly
		b := make([]byte, size)
		return &b
	}

	buf := dataPool.Get().(*[]byte)
	if cap(*buf) < size {
		// Buffer too small, allocate a new one
		*buf = make([]byte, size)
	} else {
		*buf = (*buf)[:size]
	}
	return buf
}

// ObjectPool provides methods for using pooled objects.
// This is useful for high-throughput applications that want to reduce GC pressure.
type ObjectPool struct{}

// DefaultObjectPool is the default object pool instance.
var DefaultObjectPool = &ObjectPool{}

// Get acquires a PooledObject from the pool.
// The caller must call Release() when done with the object.
func (p *ObjectPool) Get() *PooledObject {
	return acquirePooledObject()
}

// GetWithData acquires a PooledObject with a pre-allocated data buffer.
// The caller must call Release() when done with the object.
func (p *ObjectPool) GetWithData(dataSize int) *PooledObject {
	obj := acquirePooledObject()
	if dataSize > 0 {
		obj.pooledData = acquireDataBuffer(dataSize)
		obj.Object.Data = *obj.pooledData
	}
	return obj
}
