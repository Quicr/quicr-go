// SPDX-FileCopyrightText: Copyright (c) 2025 Cisco Systems
// SPDX-License-Identifier: BSD-2-Clause

package qgo

import (
	"encoding/binary"
	"testing"
)

func TestObjectHeadersExtensions(t *testing.T) {
	// Build metadata: response_type=1, flags=0x01 (has_translated)
	metadata := make([]byte, 8)
	binary.BigEndian.PutUint32(metadata[0:4], 1)
	binary.BigEndian.PutUint32(metadata[4:8], 0x01)

	headers := ObjectHeaders{
		GroupID:    10,
		ObjectID:   5,
		SubgroupID: 0,
		Priority:   128,
		TTL:        2000,
		Status:     ObjectStatusAvailable,
		Extensions: []Extension{
			{Key: 0x10, Value: metadata},
		},
	}

	if len(headers.Extensions) != 1 {
		t.Fatalf("expected 1 extension, got %d", len(headers.Extensions))
	}
	if headers.Extensions[0].Key != 0x10 {
		t.Errorf("extension key = 0x%x, want 0x10", headers.Extensions[0].Key)
	}
	if len(headers.Extensions[0].Value) != 8 {
		t.Fatalf("extension value len = %d, want 8", len(headers.Extensions[0].Value))
	}

	// Verify round-trip of metadata values
	rt := binary.BigEndian.Uint32(headers.Extensions[0].Value[0:4])
	flags := binary.BigEndian.Uint32(headers.Extensions[0].Value[4:8])
	if rt != 1 {
		t.Errorf("response_type = %d, want 1", rt)
	}
	if flags != 0x01 {
		t.Errorf("flags = 0x%x, want 0x01", flags)
	}
}

func TestObjectHeadersExtensionsCGORoundTrip(t *testing.T) {
	metadata := make([]byte, 8)
	binary.BigEndian.PutUint32(metadata[0:4], 2)
	binary.BigEndian.PutUint32(metadata[4:8], 0x03)

	headers := ObjectHeaders{
		GroupID:    42,
		ObjectID:   7,
		SubgroupID: 0,
		Priority:   64,
		TTL:        1000,
		Status:     ObjectStatusAvailable,
		Extensions: []Extension{
			{Key: 0x10, Value: metadata},
		},
	}

	// Convert Go -> C
	cH := objectHeadersToC(headers)

	// Verify C struct has extension
	if cH.num_extensions != 1 {
		t.Fatalf("C num_extensions = %d, want 1", cH.num_extensions)
	}
	if uint64(cH.extensions[0].key) != 0x10 {
		t.Errorf("C extension key = 0x%x, want 0x10", cH.extensions[0].key)
	}
	if uint16(cH.extensions[0].value_len) != 8 {
		t.Fatalf("C extension value_len = %d, want 8", cH.extensions[0].value_len)
	}
}
