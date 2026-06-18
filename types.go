// SPDX-FileCopyrightText: Copyright (c) 2025 Cisco Systems
// SPDX-License-Identifier: BSD-2-Clause

//go:generate go run generate.go

package qgo

import (
	"strings"
)

// MaxNamespaceEntries is the maximum number of entries in a namespace.
const MaxNamespaceEntries = 32

// MaxEntrySize is the maximum size of a single namespace or track name entry.
const MaxEntrySize = 256

// Namespace represents a hierarchical track namespace.
// A namespace consists of 1-32 entries, each up to 256 bytes.
type Namespace struct {
	entries [][]byte
}

// NewNamespace creates a namespace from string parts.
// Example: NewNamespace("example", "room", "123")
func NewNamespace(parts ...string) Namespace {
	entries := make([][]byte, len(parts))
	for i, p := range parts {
		entries[i] = []byte(p)
	}
	return Namespace{entries: entries}
}

// NewNamespaceFromBytes creates a namespace from byte slices.
func NewNamespaceFromBytes(entries ...[]byte) Namespace {
	copied := make([][]byte, len(entries))
	for i, e := range entries {
		copied[i] = make([]byte, len(e))
		copy(copied[i], e)
	}
	return Namespace{entries: copied}
}

// ParseNamespace parses a slash-separated namespace string.
// Example: "example/room/123" -> Namespace with 3 entries
func ParseNamespace(s string) Namespace {
	s = strings.Trim(s, "/")
	if s == "" {
		return Namespace{}
	}
	parts := strings.Split(s, "/")
	return NewNamespace(parts...)
}

// Entries returns a copy of the namespace entries.
// The returned slice can be safely modified without affecting the namespace.
func (n Namespace) Entries() [][]byte {
	if len(n.entries) == 0 {
		return nil
	}
	result := make([][]byte, len(n.entries))
	for i, e := range n.entries {
		result[i] = append([]byte(nil), e...)
	}
	return result
}

// EntriesUnsafe returns the namespace entries without copying.
// The returned slice must not be modified.
func (n Namespace) EntriesUnsafe() [][]byte {
	return n.entries
}

// NumEntries returns the number of entries in the namespace.
func (n Namespace) NumEntries() int {
	return len(n.entries)
}

// String returns a slash-separated string representation.
func (n Namespace) String() string {
	parts := make([]string, len(n.entries))
	for i, e := range n.entries {
		parts[i] = string(e)
	}
	return strings.Join(parts, "/")
}

// IsEmpty returns true if the namespace has no entries.
func (n Namespace) IsEmpty() bool {
	return len(n.entries) == 0
}

// TrackName represents a track name within a namespace.
type TrackName struct {
	data []byte
}

// NewTrackName creates a track name from a string.
func NewTrackName(name string) TrackName {
	return TrackName{data: []byte(name)}
}

// NewTrackNameFromBytes creates a track name from bytes.
func NewTrackNameFromBytes(data []byte) TrackName {
	copied := make([]byte, len(data))
	copy(copied, data)
	return TrackName{data: copied}
}

// Bytes returns a copy of the raw bytes of the track name.
// The returned slice can be safely modified without affecting the track name.
func (t TrackName) Bytes() []byte {
	if len(t.data) == 0 {
		return nil
	}
	return append([]byte(nil), t.data...)
}

// BytesUnsafe returns the raw bytes of the track name without copying.
// The returned slice must not be modified.
func (t TrackName) BytesUnsafe() []byte {
	return t.data
}

// String returns the track name as a string.
func (t TrackName) String() string {
	return string(t.data)
}

// Len returns the length of the track name.
func (t TrackName) Len() int {
	return len(t.data)
}

// FullTrackName combines a namespace and track name.
type FullTrackName struct {
	Namespace Namespace
	TrackName TrackName
}

// String returns a string representation of the full track name.
func (f FullTrackName) String() string {
	if f.Namespace.IsEmpty() {
		return f.TrackName.String()
	}
	return f.Namespace.String() + "/" + f.TrackName.String()
}

// ObjectStatus indicates the status of a published object.
type ObjectStatus uint8

const (
	// ObjectStatusAvailable indicates the object data is available.
	ObjectStatusAvailable ObjectStatus = 0
	// ObjectStatusDoesNotExist indicates the object does not exist.
	ObjectStatusDoesNotExist ObjectStatus = 1
	// ObjectStatusEndOfGroup indicates this is the last object in the group.
	ObjectStatusEndOfGroup ObjectStatus = 3
	// ObjectStatusEndOfTrack indicates this is the last object in the track.
	ObjectStatusEndOfTrack ObjectStatus = 4
	// ObjectStatusEndOfSubGroup indicates this is the last object in the subgroup.
	ObjectStatusEndOfSubGroup ObjectStatus = 5
)

// String returns a string representation of the object status.
func (s ObjectStatus) String() string {
	switch s {
	case ObjectStatusAvailable:
		return "Available"
	case ObjectStatusDoesNotExist:
		return "DoesNotExist"
	case ObjectStatusEndOfGroup:
		return "EndOfGroup"
	case ObjectStatusEndOfTrack:
		return "EndOfTrack"
	case ObjectStatusEndOfSubGroup:
		return "EndOfSubGroup"
	default:
		return "Unknown"
	}
}

// TrackMode specifies how objects are transmitted.
type TrackMode uint8

const (
	// TrackModeDatagram sends objects as QUIC datagrams.
	TrackModeDatagram TrackMode = 0
	// TrackModeStream sends objects over QUIC streams.
	TrackModeStream TrackMode = 1
)

// String returns a string representation of the track mode.
func (m TrackMode) String() string {
	switch m {
	case TrackModeDatagram:
		return "Datagram"
	case TrackModeStream:
		return "Stream"
	default:
		return "Unknown"
	}
}

// GroupOrder specifies the order in which groups are delivered.
type GroupOrder uint8

const (
	// GroupOrderOriginal delivers groups in publisher order.
	GroupOrderOriginal GroupOrder = 0
	// GroupOrderAscending delivers groups in ascending order.
	GroupOrderAscending GroupOrder = 1
	// GroupOrderDescending delivers groups in descending order.
	GroupOrderDescending GroupOrder = 2
)

// String returns a string representation of the group order.
func (o GroupOrder) String() string {
	switch o {
	case GroupOrderOriginal:
		return "Original"
	case GroupOrderAscending:
		return "Ascending"
	case GroupOrderDescending:
		return "Descending"
	default:
		return "Unknown"
	}
}

// FilterType specifies how subscription filtering is applied.
type FilterType uint8

const (
	// FilterTypeLargestObject starts from the largest available object.
	FilterTypeLargestObject FilterType = 0
	// FilterTypeLatestGroup starts from the latest group.
	FilterTypeLatestGroup FilterType = 1
	// FilterTypeLatestObject starts from the latest object.
	FilterTypeLatestObject FilterType = 2
	// FilterTypeAbsoluteStart starts from an absolute position.
	FilterTypeAbsoluteStart FilterType = 3
	// FilterTypeAbsoluteRange specifies an absolute range.
	FilterTypeAbsoluteRange FilterType = 4
)

// String returns a string representation of the filter type.
func (f FilterType) String() string {
	switch f {
	case FilterTypeLargestObject:
		return "LargestObject"
	case FilterTypeLatestGroup:
		return "LatestGroup"
	case FilterTypeLatestObject:
		return "LatestObject"
	case FilterTypeAbsoluteStart:
		return "AbsoluteStart"
	case FilterTypeAbsoluteRange:
		return "AbsoluteRange"
	default:
		return "Unknown"
	}
}

// Extension represents a single object extension key-value pair.
type Extension struct {
	Key   uint64
	Value []byte
}

// ObjectHeaders contains metadata for a published object.
type ObjectHeaders struct {
	// GroupID is the application-defined group identifier.
	GroupID uint64
	// ObjectID is the object identifier within the group.
	ObjectID uint64
	// SubgroupID is the subgroup identifier (starts at 0).
	SubgroupID uint64
	// Priority is the object priority (lower is higher priority).
	Priority uint8
	// TTL is the time-to-live in milliseconds.
	TTL uint16
	// Status is the object status.
	Status ObjectStatus
	// Extensions are optional key-value pairs attached to the object.
	Extensions []Extension
}

// Object represents a received object with headers and data.
type Object struct {
	// Headers contains the object metadata.
	Headers ObjectHeaders
	// Data contains the object payload.
	Data []byte
}

// Transport specifies the underlying transport protocol.
type Transport uint8

const (
	// TransportQUIC uses native QUIC transport.
	// Connect URI format: moq://host:port
	TransportQUIC Transport = 0
	// TransportWebTransport uses WebTransport over HTTP/3.
	// Connect URI format: https://host:port/path
	TransportWebTransport Transport = 1
)

// String returns a string representation of the transport.
func (t Transport) String() string {
	switch t {
	case TransportQUIC:
		return "QUIC"
	case TransportWebTransport:
		return "WebTransport"
	default:
		return "Unknown"
	}
}

// URIScheme returns the URI scheme for this transport.
func (t Transport) URIScheme() string {
	switch t {
	case TransportQUIC:
		return "moq"
	case TransportWebTransport:
		return "https"
	default:
		return "moq"
	}
}

// BuildURI ensures the URI has the correct scheme for this transport.
// If the URI already has a scheme, it is returned unchanged.
// Otherwise, the appropriate scheme is prepended.
func (t Transport) BuildURI(uri string) string {
	// Check if URI already has a scheme
	if strings.Contains(uri, "://") {
		return uri
	}
	return t.URIScheme() + "://" + uri
}
