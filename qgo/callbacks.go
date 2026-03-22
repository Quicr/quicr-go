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
extern void goNamespaceTrackAnnouncedCallback(quicr_full_track_name_t* full_track_name, void* user_data);
*/
import "C"

import (
	"unsafe"

	"github.com/quicr/qgo/internal/registry"
	"github.com/quicr/qgo/internal/workerpool"
)

// Global registries for callback routing.
// These map handle IDs (passed as user_data in C callbacks) to Go objects.
var (
	clientRegistry              = registry.New[*Client]()
	publishRegistry             = registry.New[*PublishTrackHandler]()
	subscribeRegistry           = registry.New[*SubscribeTrackHandler]()
	publishNamespaceRegistry    = registry.New[*PublishNamespaceHandler]()
	subscribeNamespaceRegistry  = registry.New[*SubscribeNamespaceHandler]()
)

// callbackPool is a shared worker pool for processing callbacks
// without creating a new goroutine for each callback.
var callbackPool = workerpool.New(0, 0)


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

	// Use worker pool for better performance on high-frequency callbacks
	callbackPool.Submit(func() {
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

//export goNamespaceTrackAnnouncedCallback
func goNamespaceTrackAnnouncedCallback(ftn *C.quicr_full_track_name_t, userData unsafe.Pointer) {
	handleID := uint64(uintptr(userData))
	handler, ok := subscribeNamespaceRegistry.Get(handleID)
	if !ok {
		return
	}

	handler.mu.RLock()
	callback := handler.onTrackAnnounced
	handler.mu.RUnlock()

	if callback == nil {
		return
	}

	// Convert C full track name to Go
	goFtn := convertFullTrackName(ftn)

	go callback(goFtn)
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
	C.quicr_subscribe_namespace_handler_set_track_announced_callback(
		handle,
		C.quicr_namespace_track_announced_callback_t(C.goNamespaceTrackAnnouncedCallback),
		unsafe.Pointer(uintptr(handleID)),
	)
}
