// SPDX-FileCopyrightText: Copyright (c) 2025 Cisco Systems
// SPDX-License-Identifier: BSD-2-Clause

package qgo

/*
#include "quicr_shim.h"
*/
import "C"

import (
	"context"
	"sync"
	"unsafe"
)

// PublishTrackHandler manages publishing to a track.
type PublishTrackHandler struct {
	handle   cPublishTrackHandler
	handleID uint64
	client   *Client
	config   PublishTrackConfig

	mu     sync.RWMutex
	status PublishStatus
	closed bool

	onStatusChange func(PublishStatus)
}

// newPublishTrackHandler creates a new publish track handler.
func newPublishTrackHandler(client *Client, cfg PublishTrackConfig) (*PublishTrackHandler, error) {
	var cCfg cPublishTrackConfig
	cCfg.full_track_name = fullTrackNameToC(cfg.FullTrackName)
	cCfg.track_mode = C.quicr_track_mode_t(cfg.TrackMode)
	cCfg.default_priority = C.uint8_t(cfg.Priority)
	cCfg.default_ttl = C.uint32_t(cfg.TTL)
	if cfg.UseAnnounce {
		cCfg.use_announce = 1
	} else {
		cCfg.use_announce = 0
	}

	handle := C.quicr_publish_track_handler_create(&cCfg)
	if handle == nil {
		return nil, ErrInternal
	}

	h := &PublishTrackHandler{
		handle: handle,
		client: client,
		config: cfg,
		status: PublishStatusNotConnected,
	}

	h.handleID = publishRegistry.Register(h)
	setPublishStatusCallback(handle, h.handleID)

	return h, nil
}

// CanPublish returns true if publishing is currently allowed.
func (h *PublishTrackHandler) CanPublish() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if h.closed {
		return false
	}
	return C.quicr_publish_track_handler_can_publish(h.handle) != 0
}

// Status returns the current publishing status.
func (h *PublishTrackHandler) Status() PublishStatus {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.status
}

// PublishObject publishes a single object with the given headers and data.
func (h *PublishTrackHandler) PublishObject(headers ObjectHeaders, data []byte) (PublishObjectStatus, error) {
	return h.publishObjectInternal(headers, data)
}

// PublishObjectWithContext publishes a single object with context support.
// The context is checked before attempting to publish.
func (h *PublishTrackHandler) PublishObjectWithContext(ctx context.Context, headers ObjectHeaders, data []byte) (PublishObjectStatus, error) {
	select {
	case <-ctx.Done():
		return PublishObjectInternalError, ctx.Err()
	default:
	}
	return h.publishObjectInternal(headers, data)
}

func (h *PublishTrackHandler) publishObjectInternal(headers ObjectHeaders, data []byte) (PublishObjectStatus, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if h.closed {
		return PublishObjectInternalError, ErrClosed
	}

	cHeaders := objectHeadersToC(headers)
	cHeaders.payload_length = C.uint64_t(len(data))

	var dataPtr *C.uint8_t
	if len(data) > 0 {
		dataPtr = (*C.uint8_t)(unsafe.Pointer(&data[0]))
	}

	result := C.quicr_publish_track_handler_publish_object(
		h.handle,
		&cHeaders,
		dataPtr,
		C.size_t(len(data)),
	)

	status := convertPublishObjectStatus(result)
	return status, status.Error()
}

// SetPriority updates the default publishing priority.
func (h *PublishTrackHandler) SetPriority(priority uint8) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if h.closed {
		return
	}
	C.quicr_publish_track_handler_set_priority(h.handle, C.uint8_t(priority))
}

// SetTTL updates the default TTL in milliseconds.
func (h *PublishTrackHandler) SetTTL(ttl uint32) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if h.closed {
		return
	}
	C.quicr_publish_track_handler_set_ttl(h.handle, C.uint32_t(ttl))
}

// OnStatusChange sets a callback for status changes.
// The callback is invoked in a separate goroutine.
func (h *PublishTrackHandler) OnStatusChange(fn func(PublishStatus)) {
	h.mu.Lock()
	h.onStatusChange = fn
	h.mu.Unlock()
}

// FullTrackName returns the full track name for this handler.
func (h *PublishTrackHandler) FullTrackName() FullTrackName {
	return h.config.FullTrackName
}

// Unpublish stops publishing on this track.
func (h *PublishTrackHandler) Unpublish() error {
	if h.client == nil {
		return ErrNotConnected
	}
	return h.client.UnpublishTrack(h)
}

// close releases resources without notifying the client.
func (h *PublishTrackHandler) close() {
	h.mu.Lock()
	if h.closed {
		h.mu.Unlock()
		return
	}
	h.closed = true
	h.mu.Unlock()

	publishRegistry.Unregister(h.handleID)
	C.quicr_publish_track_handler_destroy(h.handle)
	h.handle = nil
}

// Close releases all resources associated with the handler.
func (h *PublishTrackHandler) Close() error {
	if h.client != nil {
		h.client.removePublishHandler(h)
	}
	h.close()
	return nil
}
