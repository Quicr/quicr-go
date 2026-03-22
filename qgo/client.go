package qgo

/*
#include "quicr_shim.h"
#include <stdlib.h>
#include <string.h>
*/
import "C"

import (
	"context"
	"sync"
	"unsafe"
)

// Client manages a connection to a MoQ relay server.
type Client struct {
	handle   cClient
	handleID uint64
	config   ClientConfig

	mu         sync.RWMutex
	status     ClientStatus
	closed     bool
	statusCond *sync.Cond // Condition variable for status changes

	// Callbacks
	onStatusChange func(ClientStatus)

	// Track handlers
	publishHandlers   map[*PublishTrackHandler]struct{}
	subscribeHandlers map[*SubscribeTrackHandler]struct{}
	handlersMu        sync.Mutex
}

// NewClient creates a new MoQ client.
func NewClient(cfg ClientConfig) (*Client, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	// Convert Go config to C config
	var cCfg cClientConfig
	C.quicr_client_config_init(&cCfg)

	// Set endpoint ID
	endpointID := C.CString(cfg.EndpointID)
	defer C.free(unsafe.Pointer(endpointID))
	C.strncpy(&cCfg.endpoint_id[0], endpointID, C.QUICR_MAX_ENDPOINT_ID_SIZE-1)

	// Set connect URI
	connectURI := C.CString(cfg.ConnectURI)
	defer C.free(unsafe.Pointer(connectURI))
	C.strncpy(&cCfg.connect_uri[0], connectURI, C.QUICR_MAX_URI_SIZE-1)

	// Set timing parameters
	cCfg.metrics_sample_ms = C.uint64_t(cfg.MetricsSampleInterval.Milliseconds())
	cCfg.tick_service_sleep_delay_us = C.uint64_t(cfg.TickServiceDelay.Microseconds())

	// Create C client
	handle := C.quicr_client_create(&cCfg)
	if handle == nil {
		return nil, ErrInternal
	}

	c := &Client{
		handle:            handle,
		config:            cfg,
		status:            ClientStatusNotConnected,
		publishHandlers:   make(map[*PublishTrackHandler]struct{}),
		subscribeHandlers: make(map[*SubscribeTrackHandler]struct{}),
	}
	c.statusCond = sync.NewCond(&c.mu)

	// Register for callbacks
	c.handleID = clientRegistry.Register(c)
	setClientStatusCallback(handle, c.handleID)

	return c, nil
}

// Connect initiates a connection to the relay server.
// It blocks until the connection is established or the context is cancelled.
func (c *Client) Connect(ctx context.Context) error {
	c.mu.RLock()
	if c.closed {
		c.mu.RUnlock()
		return ErrClosed
	}
	c.mu.RUnlock()

	result := C.quicr_client_connect(c.handle)
	if err := resultToError(result); err != nil {
		return err
	}

	// Wait for connection or context cancellation
	return c.waitForStatus(ctx, ClientStatusReady)
}

// waitForStatus blocks until the client reaches the target status or an error occurs.
func (c *Client) waitForStatus(ctx context.Context, target ClientStatus) error {
	// Create a done channel to signal context cancellation
	done := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			c.mu.Lock()
			c.statusCond.Broadcast()
			c.mu.Unlock()
		case <-done:
		}
	}()
	defer close(done)

	c.mu.Lock()
	defer c.mu.Unlock()

	for {
		// Check context first
		if ctx.Err() != nil {
			return ctx.Err()
		}

		if c.closed {
			return ErrClosed
		}
		if c.status == target {
			return nil
		}
		if c.status == ClientStatusFailedToConnect || c.status == ClientStatusInternalError {
			return ErrNotReady
		}

		// Wait for status change signal
		c.statusCond.Wait()
	}
}

// Disconnect gracefully disconnects from the relay server.
func (c *Client) Disconnect() error {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return ErrClosed
	}
	c.mu.Unlock()

	result := C.quicr_client_disconnect(c.handle)
	return resultToError(result)
}

// Close releases all resources associated with the client.
// It disconnects from the server and cleans up all handlers.
func (c *Client) Close() error {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return nil
	}
	c.closed = true
	c.mu.Unlock()

	// Unregister from callback registry first
	clientRegistry.Unregister(c.handleID)

	// Clean up handlers
	c.handlersMu.Lock()
	for h := range c.publishHandlers {
		h.close()
	}
	for h := range c.subscribeHandlers {
		h.close()
	}
	c.publishHandlers = nil
	c.subscribeHandlers = nil
	c.handlersMu.Unlock()

	// Disconnect and destroy
	C.quicr_client_disconnect(c.handle)
	C.quicr_client_destroy(c.handle)
	c.handle = nil

	return nil
}

// Status returns the current connection status.
func (c *Client) Status() ClientStatus {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.status
}

// IsConnected returns true if the client is connected and ready.
func (c *Client) IsConnected() bool {
	return c.Status() == ClientStatusReady
}

// OnStatusChange sets a callback for connection status changes.
// The callback is invoked in a separate goroutine.
func (c *Client) OnStatusChange(fn func(ClientStatus)) {
	c.mu.Lock()
	c.onStatusChange = fn
	c.mu.Unlock()
}

// PublishNamespace announces a namespace for publishing using a namespace handler.
func (c *Client) PublishNamespace(handler *PublishNamespaceHandler) error {
	c.mu.RLock()
	if c.closed {
		c.mu.RUnlock()
		return ErrClosed
	}
	c.mu.RUnlock()

	result := C.quicr_client_publish_namespace(c.handle, handler.handle)
	return resultToError(result)
}

// PublishNamespaceDone signals that publishing to a namespace is complete.
func (c *Client) PublishNamespaceDone(handler *PublishNamespaceHandler) error {
	c.mu.RLock()
	if c.closed {
		c.mu.RUnlock()
		return ErrClosed
	}
	c.mu.RUnlock()

	result := C.quicr_client_publish_namespace_done(c.handle, handler.handle)
	return resultToError(result)
}

// SubscribeNamespace subscribes to a namespace prefix to receive track announcements.
func (c *Client) SubscribeNamespace(handler *SubscribeNamespaceHandler) error {
	c.mu.RLock()
	if c.closed {
		c.mu.RUnlock()
		return ErrClosed
	}
	c.mu.RUnlock()

	result := C.quicr_client_subscribe_namespace(c.handle, handler.handle)
	return resultToError(result)
}

// UnsubscribeNamespace unsubscribes from a namespace prefix.
func (c *Client) UnsubscribeNamespace(handler *SubscribeNamespaceHandler) error {
	c.mu.RLock()
	if c.closed {
		c.mu.RUnlock()
		return ErrClosed
	}
	c.mu.RUnlock()

	result := C.quicr_client_unsubscribe_namespace(c.handle, handler.handle)
	return resultToError(result)
}

// PublishTrack creates and registers a publish track handler.
func (c *Client) PublishTrack(cfg PublishTrackConfig) (*PublishTrackHandler, error) {
	c.mu.RLock()
	if c.closed {
		c.mu.RUnlock()
		return nil, ErrClosed
	}
	c.mu.RUnlock()

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	handler, err := newPublishTrackHandler(c, cfg)
	if err != nil {
		return nil, err
	}

	c.handlersMu.Lock()
	c.publishHandlers[handler] = struct{}{}
	c.handlersMu.Unlock()

	result := C.quicr_client_publish_track(c.handle, handler.handle)
	if err := resultToError(result); err != nil {
		c.handlersMu.Lock()
		delete(c.publishHandlers, handler)
		c.handlersMu.Unlock()
		handler.close()
		return nil, err
	}

	return handler, nil
}

// UnpublishTrack removes a publish track handler.
func (c *Client) UnpublishTrack(handler *PublishTrackHandler) error {
	c.mu.RLock()
	if c.closed {
		c.mu.RUnlock()
		return ErrClosed
	}
	c.mu.RUnlock()

	result := C.quicr_client_unpublish_track(c.handle, handler.handle)
	if err := resultToError(result); err != nil {
		return err
	}

	c.handlersMu.Lock()
	delete(c.publishHandlers, handler)
	c.handlersMu.Unlock()

	return nil
}

// SubscribeTrack creates and registers a subscribe track handler.
func (c *Client) SubscribeTrack(cfg SubscribeTrackConfig) (*SubscribeTrackHandler, error) {
	c.mu.RLock()
	if c.closed {
		c.mu.RUnlock()
		return nil, ErrClosed
	}
	c.mu.RUnlock()

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	handler, err := newSubscribeTrackHandler(c, cfg)
	if err != nil {
		return nil, err
	}

	c.handlersMu.Lock()
	c.subscribeHandlers[handler] = struct{}{}
	c.handlersMu.Unlock()

	result := C.quicr_client_subscribe_track(c.handle, handler.handle)
	if err := resultToError(result); err != nil {
		c.handlersMu.Lock()
		delete(c.subscribeHandlers, handler)
		c.handlersMu.Unlock()
		handler.close()
		return nil, err
	}

	return handler, nil
}

// UnsubscribeTrack removes a subscribe track handler.
func (c *Client) UnsubscribeTrack(handler *SubscribeTrackHandler) error {
	c.mu.RLock()
	if c.closed {
		c.mu.RUnlock()
		return ErrClosed
	}
	c.mu.RUnlock()

	result := C.quicr_client_unsubscribe_track(c.handle, handler.handle)
	if err := resultToError(result); err != nil {
		return err
	}

	c.handlersMu.Lock()
	delete(c.subscribeHandlers, handler)
	c.handlersMu.Unlock()

	return nil
}

// removePublishHandler removes a handler from the client's internal tracking.
func (c *Client) removePublishHandler(handler *PublishTrackHandler) {
	c.handlersMu.Lock()
	delete(c.publishHandlers, handler)
	c.handlersMu.Unlock()
}

// removeSubscribeHandler removes a handler from the client's internal tracking.
func (c *Client) removeSubscribeHandler(handler *SubscribeTrackHandler) {
	c.handlersMu.Lock()
	delete(c.subscribeHandlers, handler)
	c.handlersMu.Unlock()
}
