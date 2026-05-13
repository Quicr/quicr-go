// SPDX-FileCopyrightText: Copyright (c) 2025 Cisco Systems
// SPDX-License-Identifier: BSD-2-Clause

package qgo

/*
#include "quicr_shim.h"
*/
import "C"

import (
	"sync"
)

// SubscribeTrackHandler manages subscription to a track.
type SubscribeTrackHandler struct {
	handle   cSubscribeTrackHandler
	handleID uint64
	client   *Client
	config   SubscribeTrackConfig

	mu     sync.RWMutex
	status SubscribeStatus
	closed bool

	onObjectReceived func(Object)
	onStatusChange   func(SubscribeStatus)
}

// newSubscribeTrackHandler creates a new subscribe track handler.
func newSubscribeTrackHandler(client *Client, cfg SubscribeTrackConfig) (*SubscribeTrackHandler, error) {
	var cCfg cSubscribeTrackConfig
	cCfg.full_track_name = fullTrackNameToC(cfg.FullTrackName)
	cCfg.priority = C.uint8_t(cfg.Priority)
	cCfg.group_order = C.quicr_group_order_t(cfg.GroupOrder)
	cCfg.filter_type = C.quicr_filter_type_t(cfg.FilterType)

	handle := C.quicr_subscribe_track_handler_create(&cCfg)
	if handle == nil {
		return nil, ErrInternal
	}

	h := &SubscribeTrackHandler{
		handle: handle,
		client: client,
		config: cfg,
		status: SubscribeStatusNotSubscribed,
	}

	h.handleID = subscribeRegistry.Register(h)
	setSubscribeCallbacks(handle, h.handleID)

	return h, nil
}

// Status returns the current subscription status.
func (h *SubscribeTrackHandler) Status() SubscribeStatus {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.status
}

// IsActive returns true if the subscription is currently receiving data.
func (h *SubscribeTrackHandler) IsActive() bool {
	return h.Status() == SubscribeStatusOK
}

// OnObjectReceived sets the callback for received objects.
// The callback is invoked in a separate goroutine for each object.
func (h *SubscribeTrackHandler) OnObjectReceived(fn func(Object)) {
	h.mu.Lock()
	h.onObjectReceived = fn
	h.mu.Unlock()
}

// OnStatusChange sets a callback for status changes.
// The callback is invoked in a separate goroutine.
func (h *SubscribeTrackHandler) OnStatusChange(fn func(SubscribeStatus)) {
	h.mu.Lock()
	h.onStatusChange = fn
	h.mu.Unlock()
}

// SetPriority updates the subscription priority.
func (h *SubscribeTrackHandler) SetPriority(priority uint8) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if h.closed {
		return
	}
	C.quicr_subscribe_track_handler_set_priority(h.handle, C.uint8_t(priority))
}

// Pause temporarily stops receiving objects.
func (h *SubscribeTrackHandler) Pause() error {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if h.closed {
		return ErrClosed
	}
	result := C.quicr_subscribe_track_handler_pause(h.handle)
	return resultToError(result)
}

// Resume resumes receiving objects after pause.
func (h *SubscribeTrackHandler) Resume() error {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if h.closed {
		return ErrClosed
	}
	result := C.quicr_subscribe_track_handler_resume(h.handle)
	return resultToError(result)
}

// FullTrackName returns the full track name for this handler.
func (h *SubscribeTrackHandler) FullTrackName() FullTrackName {
	return h.config.FullTrackName
}

// Unsubscribe stops the subscription.
func (h *SubscribeTrackHandler) Unsubscribe() error {
	if h.client == nil {
		return ErrNotConnected
	}
	return h.client.UnsubscribeTrack(h)
}

// close releases resources without notifying the client.
func (h *SubscribeTrackHandler) close() {
	h.mu.Lock()
	if h.closed {
		h.mu.Unlock()
		return
	}
	h.closed = true
	h.mu.Unlock()

	subscribeRegistry.Unregister(h.handleID)
	C.quicr_subscribe_track_handler_destroy(h.handle)
	h.handle = nil
}

// Close releases all resources associated with the handler.
func (h *SubscribeTrackHandler) Close() error {
	if h.client != nil {
		h.client.removeSubscribeHandler(h)
	}
	h.close()
	return nil
}
