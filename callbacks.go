// SPDX-FileCopyrightText: Copyright (c) 2025 Cisco Systems
// SPDX-License-Identifier: BSD-2-Clause

package qgo

/*
#include "quicr_shim.h"

// Forward declarations for Go-exported callback functions
extern void goClientStatusCallback(quicr_client_status_t status, void* user_data);
extern void goPublishStatusCallback(quicr_publish_status_t status, void* user_data);
extern void goSubscribeStatusCallback(quicr_subscribe_status_t status, void* user_data);
extern void goObjectReceivedCallback(quicr_object_t* object, void* user_data);
extern void goPublishNamespaceStatusCallback(quicr_publish_namespace_status_t status, void* user_data);
extern void goSubscribeNamespaceStatusCallback(quicr_subscribe_namespace_status_t status, void* user_data);
extern void goPublishNamespaceReceivedCallback(quicr_namespace_t* track_namespace, void* user_data);
extern void goPublishReceivedCallback(quicr_full_track_name_t* full_track_name, uint64_t track_alias, void* user_data);
*/
import "C"

import (
	"sync"
	"unsafe"

	"github.com/quicr/qgo/internal/registry"
	"github.com/quicr/qgo/internal/workerpool"
)

// Global registries for callback routing.
// These map handle IDs (passed as user_data in C callbacks) to Go objects.
var (
	clientRegistry             = registry.New[*Client]()
	publishRegistry            = registry.New[*PublishTrackHandler]()
	subscribeRegistry          = registry.New[*SubscribeTrackHandler]()
	publishNamespaceRegistry   = registry.New[*PublishNamespaceHandler]()
	subscribeNamespaceRegistry = registry.New[*SubscribeNamespaceHandler]()
)

// callbackPool is a shared striped worker pool for processing callbacks.
// Using a striped pool reduces channel contention under high parallelism.
var callbackPool = workerpool.NewStriped(0, 0, 0)

// callbackPoolMu protects callbackPool during reconfiguration.
var callbackPoolMu sync.RWMutex

// CallbackPoolConfig configures the callback worker pool.
type CallbackPoolConfig struct {
	// Stripes is the number of independent worker pools.
	// Default: runtime.NumCPU() (capped at 4-16)
	Stripes int

	// WorkersPerStripe is the number of workers per stripe.
	// Default: 2
	WorkersPerStripe int

	// QueueSizePerStripe is the queue size per stripe.
	// Default: WorkersPerStripe * 128
	QueueSizePerStripe int
}

// ConfigureCallbackPool reconfigures the global callback pool.
// This should be called before creating any clients.
// It closes the existing pool and creates a new one with the given config.
func ConfigureCallbackPool(cfg CallbackPoolConfig) {
	callbackPoolMu.Lock()
	defer callbackPoolMu.Unlock()

	// Close existing pool
	callbackPool.Close()

	// Create new pool with config
	callbackPool = workerpool.NewStriped(
		cfg.Stripes,
		cfg.WorkersPerStripe,
		cfg.QueueSizePerStripe,
	)
}


//export goClientStatusCallback
func goClientStatusCallback(status C.quicr_client_status_t, userData unsafe.Pointer) {
	handleID := uint64(uintptr(userData))
	client, ok := clientRegistry.Get(handleID)
	if !ok {
		return // Client was destroyed
	}

	goStatus := convertClientStatus(status)

	client.mu.Lock()
	client.status = goStatus
	// Signal waiters that status changed
	client.statusCond.Broadcast()
	callback := client.onStatusChange
	client.mu.Unlock()

	// Invoke user callback in goroutine to avoid blocking C thread
	if callback != nil {
		go callback(goStatus)
	}
}

//export goPublishStatusCallback
func goPublishStatusCallback(status C.quicr_publish_status_t, userData unsafe.Pointer) {
	handleID := uint64(uintptr(userData))
	handler, ok := publishRegistry.Get(handleID)
	if !ok {
		return // Handler was destroyed
	}

	goStatus := convertPublishStatus(status)

	handler.mu.Lock()
	handler.status = goStatus
	callback := handler.onStatusChange
	handler.mu.Unlock()

	// Invoke user callback in goroutine
	if callback != nil {
		go callback(goStatus)
	}
}

//export goSubscribeStatusCallback
func goSubscribeStatusCallback(status C.quicr_subscribe_status_t, userData unsafe.Pointer) {
	handleID := uint64(uintptr(userData))
	handler, ok := subscribeRegistry.Get(handleID)
	if !ok {
		return // Handler was destroyed
	}

	goStatus := convertSubscribeStatus(status)

	handler.mu.Lock()
	handler.status = goStatus
	callback := handler.onStatusChange
	handler.mu.Unlock()

	// Invoke user callback in goroutine
	if callback != nil {
		go callback(goStatus)
	}
}

//export goObjectReceivedCallback
func goObjectReceivedCallback(obj *C.quicr_object_t, userData unsafe.Pointer) {
	handleID := uint64(uintptr(userData))
	handler, ok := subscribeRegistry.Get(handleID)
	if !ok {
		return // Handler was destroyed
	}

	// Get callback reference under lock to avoid races
	handler.mu.RLock()
	callback := handler.onObjectReceived
	handler.mu.RUnlock()

	// Early exit if no callback - avoid unnecessary data conversion
	if callback == nil {
		return
	}

	// Convert C object to Go object (copies data to avoid dangling pointer)
	goObj := convertObject(obj)

	// Use striped worker pool for better performance on high-frequency callbacks.
	// Submit to a stripe based on handle ID to maintain ordering per subscription.
	callbackPoolMu.RLock()
	pool := callbackPool
	callbackPoolMu.RUnlock()
	pool.SubmitTo(handleID, func() {
		callback(goObj)
	})
}

// setClientStatusCallback sets the C callback for a client.
func setClientStatusCallback(handle cClient, handleID uint64) {
	C.quicr_client_set_status_callback(
		handle,
		C.quicr_client_status_callback_t(C.goClientStatusCallback),
		unsafe.Pointer(uintptr(handleID)),
	)
}

// setClientPublishNamespaceReceivedCallback sets the callback for namespace announcements (ANNOUNCE flow).
func setClientPublishNamespaceReceivedCallback(handle cClient, handleID uint64) {
	C.quicr_client_set_publish_namespace_received_callback(
		handle,
		C.quicr_namespace_track_announced_callback_t(C.goPublishNamespaceReceivedCallback),
		unsafe.Pointer(uintptr(handleID)),
	)
}

// setClientPublishReceivedCallback sets the callback for PUBLISH messages (SubNS flow).
func setClientPublishReceivedCallback(handle cClient, handleID uint64) {
	C.quicr_client_set_publish_received_callback(
		handle,
		C.quicr_publish_received_callback_t(C.goPublishReceivedCallback),
		unsafe.Pointer(uintptr(handleID)),
	)
}

// setPublishStatusCallback sets the C callback for a publish handler.
func setPublishStatusCallback(handle cPublishTrackHandler, handleID uint64) {
	C.quicr_publish_track_handler_set_status_callback(
		handle,
		C.quicr_publish_status_callback_t(C.goPublishStatusCallback),
		unsafe.Pointer(uintptr(handleID)),
	)
}

// setSubscribeCallbacks sets the C callbacks for a subscribe handler.
func setSubscribeCallbacks(handle cSubscribeTrackHandler, handleID uint64) {
	C.quicr_subscribe_track_handler_set_status_callback(
		handle,
		C.quicr_subscribe_status_callback_t(C.goSubscribeStatusCallback),
		unsafe.Pointer(uintptr(handleID)),
	)
	C.quicr_subscribe_track_handler_set_object_callback(
		handle,
		C.quicr_object_received_callback_t(C.goObjectReceivedCallback),
		unsafe.Pointer(uintptr(handleID)),
	)
}

//export goPublishNamespaceStatusCallback
func goPublishNamespaceStatusCallback(status C.quicr_publish_namespace_status_t, userData unsafe.Pointer) {
	handleID := uint64(uintptr(userData))
	handler, ok := publishNamespaceRegistry.Get(handleID)
	if !ok {
		return
	}

	goStatus := convertPublishNamespaceStatus(status)

	handler.mu.Lock()
	handler.status = goStatus
	callback := handler.onStatusChange
	handler.mu.Unlock()

	if callback != nil {
		go callback(goStatus)
	}
}

//export goSubscribeNamespaceStatusCallback
func goSubscribeNamespaceStatusCallback(status C.quicr_subscribe_namespace_status_t, userData unsafe.Pointer) {
	handleID := uint64(uintptr(userData))
	handler, ok := subscribeNamespaceRegistry.Get(handleID)
	if !ok {
		return
	}

	goStatus := convertSubscribeNamespaceStatus(status)

	handler.mu.Lock()
	handler.status = goStatus
	callback := handler.onStatusChange
	handler.mu.Unlock()

	if callback != nil {
		go callback(goStatus)
	}
}

//export goPublishNamespaceReceivedCallback
func goPublishNamespaceReceivedCallback(ns *C.quicr_namespace_t, userData unsafe.Pointer) {
	handleID := uint64(uintptr(userData))
	client, ok := clientRegistry.Get(handleID)
	if !ok {
		return
	}

	client.mu.RLock()
	callback := client.onPublishNamespaceReceived
	client.mu.RUnlock()

	if callback == nil {
		return
	}

	// Convert C namespace to Go
	goNs := convertNamespace(ns)

	go callback(goNs)
}

//export goPublishReceivedCallback
func goPublishReceivedCallback(ftn *C.quicr_full_track_name_t, trackAlias C.uint64_t, userData unsafe.Pointer) {
	handleID := uint64(uintptr(userData))
	client, ok := clientRegistry.Get(handleID)
	if !ok {
		return
	}

	client.mu.RLock()
	callback := client.onPublishReceived
	client.mu.RUnlock()

	if callback == nil {
		return
	}

	// Convert C full track name to Go
	goFtn := convertFullTrackName(ftn)

	go callback(goFtn, uint64(trackAlias))
}

// setPublishNamespaceStatusCallback sets the C callback for a publish namespace handler.
func setPublishNamespaceStatusCallback(handle cPublishNamespaceHandler, handleID uint64) {
	C.quicr_publish_namespace_handler_set_status_callback(
		handle,
		C.quicr_publish_namespace_status_callback_t(C.goPublishNamespaceStatusCallback),
		unsafe.Pointer(uintptr(handleID)),
	)
}

// setSubscribeNamespaceCallbacks sets the C callbacks for a subscribe namespace handler.
func setSubscribeNamespaceCallbacks(handle cSubscribeNamespaceHandler, handleID uint64) {
	C.quicr_subscribe_namespace_handler_set_status_callback(
		handle,
		C.quicr_subscribe_namespace_status_callback_t(C.goSubscribeNamespaceStatusCallback),
		unsafe.Pointer(uintptr(handleID)),
	)
}
