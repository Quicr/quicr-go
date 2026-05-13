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

// PublishNamespaceStatus represents the status of a publish namespace operation.
type PublishNamespaceStatus uint8

const (
	PublishNamespaceStatusOK PublishNamespaceStatus = iota
	PublishNamespaceStatusNotConnected
	PublishNamespaceStatusNotPublished
	PublishNamespaceStatusPendingResponse
	PublishNamespaceStatusNotAuthorized
	PublishNamespaceStatusSendingDone
	PublishNamespaceStatusError
)

// String returns a string representation of the status.
func (s PublishNamespaceStatus) String() string {
	switch s {
	case PublishNamespaceStatusOK:
		return "OK"
	case PublishNamespaceStatusNotConnected:
		return "NotConnected"
	case PublishNamespaceStatusNotPublished:
		return "NotPublished"
	case PublishNamespaceStatusPendingResponse:
		return "PendingResponse"
	case PublishNamespaceStatusNotAuthorized:
		return "NotAuthorized"
	case PublishNamespaceStatusSendingDone:
		return "SendingDone"
	case PublishNamespaceStatusError:
		return "Error"
	default:
		return "Unknown"
	}
}

// SubscribeNamespaceStatus represents the status of a subscribe namespace operation.
type SubscribeNamespaceStatus uint8

const (
	SubscribeNamespaceStatusOK SubscribeNamespaceStatus = iota
	SubscribeNamespaceStatusNotSubscribed
	SubscribeNamespaceStatusError
)

// String returns a string representation of the status.
func (s SubscribeNamespaceStatus) String() string {
	switch s {
	case SubscribeNamespaceStatusOK:
		return "OK"
	case SubscribeNamespaceStatusNotSubscribed:
		return "NotSubscribed"
	case SubscribeNamespaceStatusError:
		return "Error"
	default:
		return "Unknown"
	}
}

// PublishNamespaceHandler manages publishing to a namespace.
type PublishNamespaceHandler struct {
	handle   cPublishNamespaceHandler
	handleID uint64
	prefix   Namespace

	mu     sync.RWMutex
	status PublishNamespaceStatus
	closed bool

	onStatusChange func(PublishNamespaceStatus)
}

// NewPublishNamespaceHandler creates a new publish namespace handler.
func NewPublishNamespaceHandler(prefix Namespace) (*PublishNamespaceHandler, error) {
	cNs := namespaceToC(prefix)
	handle := C.quicr_publish_namespace_handler_create(&cNs)
	if handle == nil {
		return nil, ErrInternal
	}

	h := &PublishNamespaceHandler{
		handle: handle,
		prefix: prefix,
		status: PublishNamespaceStatusNotPublished,
	}

	h.handleID = publishNamespaceRegistry.Register(h)
	setPublishNamespaceStatusCallback(handle, h.handleID)

	return h, nil
}

// Status returns the current status.
func (h *PublishNamespaceHandler) Status() PublishNamespaceStatus {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.status
}

// Prefix returns the namespace prefix.
func (h *PublishNamespaceHandler) Prefix() Namespace {
	return h.prefix
}

// OnStatusChange sets a callback for status changes.
func (h *PublishNamespaceHandler) OnStatusChange(fn func(PublishNamespaceStatus)) {
	h.mu.Lock()
	h.onStatusChange = fn
	h.mu.Unlock()
}

// Close releases all resources.
func (h *PublishNamespaceHandler) Close() error {
	h.mu.Lock()
	if h.closed {
		h.mu.Unlock()
		return nil
	}
	h.closed = true
	h.mu.Unlock()

	publishNamespaceRegistry.Unregister(h.handleID)
	C.quicr_publish_namespace_handler_destroy(h.handle)
	h.handle = nil

	return nil
}

// SubscribeNamespaceHandler manages subscription to a namespace prefix.
type SubscribeNamespaceHandler struct {
	handle   cSubscribeNamespaceHandler
	handleID uint64
	prefix   Namespace

	mu     sync.RWMutex
	status SubscribeNamespaceStatus
	closed bool

	onStatusChange func(SubscribeNamespaceStatus)
}

// NewSubscribeNamespaceHandler creates a new subscribe namespace handler.
func NewSubscribeNamespaceHandler(prefix Namespace) (*SubscribeNamespaceHandler, error) {
	cNs := namespaceToC(prefix)
	handle := C.quicr_subscribe_namespace_handler_create(&cNs)
	if handle == nil {
		return nil, ErrInternal
	}

	h := &SubscribeNamespaceHandler{
		handle: handle,
		prefix: prefix,
		status: SubscribeNamespaceStatusNotSubscribed,
	}

	h.handleID = subscribeNamespaceRegistry.Register(h)
	setSubscribeNamespaceCallbacks(handle, h.handleID)

	return h, nil
}

// Status returns the current status.
func (h *SubscribeNamespaceHandler) Status() SubscribeNamespaceStatus {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.status
}

// Prefix returns the namespace prefix.
func (h *SubscribeNamespaceHandler) Prefix() Namespace {
	return h.prefix
}

// OnStatusChange sets a callback for status changes.
func (h *SubscribeNamespaceHandler) OnStatusChange(fn func(SubscribeNamespaceStatus)) {
	h.mu.Lock()
	h.onStatusChange = fn
	h.mu.Unlock()
}

// Close releases all resources.
func (h *SubscribeNamespaceHandler) Close() error {
	h.mu.Lock()
	if h.closed {
		h.mu.Unlock()
		return nil
	}
	h.closed = true
	h.mu.Unlock()

	subscribeNamespaceRegistry.Unregister(h.handleID)
	C.quicr_subscribe_namespace_handler_destroy(h.handle)
	h.handle = nil

	return nil
}

// convertPublishNamespaceStatus converts C status to Go.
func convertPublishNamespaceStatus(status cPublishNamespaceStatus) PublishNamespaceStatus {
	switch status {
	case C.QUICR_PUBLISH_NAMESPACE_STATUS_OK:
		return PublishNamespaceStatusOK
	case C.QUICR_PUBLISH_NAMESPACE_STATUS_NOT_CONNECTED:
		return PublishNamespaceStatusNotConnected
	case C.QUICR_PUBLISH_NAMESPACE_STATUS_NOT_PUBLISHED:
		return PublishNamespaceStatusNotPublished
	case C.QUICR_PUBLISH_NAMESPACE_STATUS_PENDING_RESPONSE:
		return PublishNamespaceStatusPendingResponse
	case C.QUICR_PUBLISH_NAMESPACE_STATUS_NOT_AUTHORIZED:
		return PublishNamespaceStatusNotAuthorized
	case C.QUICR_PUBLISH_NAMESPACE_STATUS_SENDING_DONE:
		return PublishNamespaceStatusSendingDone
	case C.QUICR_PUBLISH_NAMESPACE_STATUS_ERROR:
		return PublishNamespaceStatusError
	default:
		return PublishNamespaceStatusError
	}
}

// convertSubscribeNamespaceStatus converts C status to Go.
func convertSubscribeNamespaceStatus(status cSubscribeNamespaceStatus) SubscribeNamespaceStatus {
	switch status {
	case C.QUICR_SUBSCRIBE_NAMESPACE_STATUS_OK:
		return SubscribeNamespaceStatusOK
	case C.QUICR_SUBSCRIBE_NAMESPACE_STATUS_NOT_SUBSCRIBED:
		return SubscribeNamespaceStatusNotSubscribed
	case C.QUICR_SUBSCRIBE_NAMESPACE_STATUS_ERROR:
		return SubscribeNamespaceStatusError
	default:
		return SubscribeNamespaceStatusError
	}
}
