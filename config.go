// SPDX-FileCopyrightText: Copyright (c) 2025 Cisco Systems
// SPDX-License-Identifier: BSD-2-Clause

package qgo

import "time"

// DefaultMetricsSampleInterval is the default interval for metrics sampling.
const DefaultMetricsSampleInterval = 5 * time.Second

// DefaultTickServiceDelay is the default tick service delay.
const DefaultTickServiceDelay = 333 * time.Microsecond

// TLSConfig configures TLS settings for the connection.
type TLSConfig struct {
	// InsecureSkipVerify disables server certificate verification.
	// WARNING: This should only be used for testing.
	InsecureSkipVerify bool

	// CertFile is the path to the client certificate file (PEM format).
	// Used for mutual TLS authentication.
	CertFile string

	// KeyFile is the path to the client private key file (PEM format).
	// Used for mutual TLS authentication.
	KeyFile string

	// CAFile is the path to a custom CA certificate file (PEM format).
	// If set, only certificates signed by this CA will be accepted.
	CAFile string
}

// WorkerPoolConfig configures the callback worker pool.
// This affects how received objects are dispatched to callbacks.
type WorkerPoolConfig struct {
	// Stripes is the number of independent worker pools.
	// More stripes reduce contention but use more memory.
	// Default: runtime.NumCPU() (capped at 4-16)
	Stripes int

	// WorkersPerStripe is the number of workers per stripe.
	// Default: 2
	WorkersPerStripe int

	// QueueSizePerStripe is the queue size per stripe.
	// Larger queues absorb bursts better but use more memory.
	// Default: WorkersPerStripe * 128
	QueueSizePerStripe int
}

// ClientConfig configures a quicr Client.
type ClientConfig struct {
	// ConnectURI is the relay server URI.
	// The URI scheme depends on the Transport setting:
	//   - QUIC: moq://hostname[:port]
	//   - WebTransport: https://hostname[:port][/path]
	// If the scheme is omitted, it will be derived from the Transport setting.
	ConnectURI string

	// EndpointID uniquely identifies this client.
	// If empty, a default will be used.
	EndpointID string

	// Transport specifies the underlying transport protocol.
	// Default: TransportQUIC.
	Transport Transport

	// TLS configures TLS settings.
	// If nil, default TLS settings are used.
	TLS *TLSConfig

	// WorkerPool configures the callback worker pool.
	// If nil, default settings are used.
	// Note: This is a global setting; only the first client's config is applied.
	WorkerPool *WorkerPoolConfig

	// MetricsSampleInterval controls how often metrics are sampled.
	// Default: 5 seconds.
	MetricsSampleInterval time.Duration

	// TickServiceDelay controls the internal tick service delay.
	// Default: 333 microseconds.
	TickServiceDelay time.Duration
}

// Validate validates the client configuration and applies defaults.
func (c *ClientConfig) Validate() error {
	if c.ConnectURI == "" {
		return ErrInvalidParam
	}
	if c.EndpointID == "" {
		c.EndpointID = "qgo-client"
	}
	if c.MetricsSampleInterval == 0 {
		c.MetricsSampleInterval = DefaultMetricsSampleInterval
	}
	if c.TickServiceDelay == 0 {
		c.TickServiceDelay = DefaultTickServiceDelay
	}
	// Add URI scheme if missing based on transport
	c.ConnectURI = c.Transport.BuildURI(c.ConnectURI)
	return nil
}

// PublishTrackConfig configures track publishing.
type PublishTrackConfig struct {
	// FullTrackName is the full name of the track to publish.
	FullTrackName FullTrackName

	// TrackMode specifies how objects are transmitted.
	// Default: TrackModeStream.
	TrackMode TrackMode

	// Priority is the default priority for published objects.
	// Lower values indicate higher priority.
	// Default: 128.
	Priority uint8

	// TTL is the default time-to-live in milliseconds.
	// Default: 5000 (5 seconds).
	TTL uint32

	// UseAnnounce specifies whether to use the announce flow.
	// If false, the publish flow is used.
	// Default: false.
	UseAnnounce bool
}

// Validate validates the publish track configuration and applies defaults.
func (c *PublishTrackConfig) Validate() error {
	if c.FullTrackName.Namespace.IsEmpty() {
		return ErrInvalidParam
	}
	if c.Priority == 0 {
		c.Priority = 128
	}
	if c.TTL == 0 {
		c.TTL = 5000
	}
	return nil
}

// SubscribeTrackConfig configures track subscription.
type SubscribeTrackConfig struct {
	// FullTrackName is the full name of the track to subscribe to.
	FullTrackName FullTrackName

	// Priority is the subscription priority.
	// Lower values indicate higher priority.
	// Default: 128.
	Priority uint8

	// GroupOrder specifies the order for group delivery.
	// Default: GroupOrderAscending.
	GroupOrder GroupOrder

	// FilterType specifies the subscription filter type.
	// Default: FilterTypeLargestObject.
	FilterType FilterType
}

// Validate validates the subscribe track configuration and applies defaults.
func (c *SubscribeTrackConfig) Validate() error {
	if c.FullTrackName.Namespace.IsEmpty() {
		return ErrInvalidParam
	}
	if c.Priority == 0 {
		c.Priority = 128
	}
	return nil
}
