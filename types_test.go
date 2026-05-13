// SPDX-FileCopyrightText: Copyright (c) 2025 Cisco Systems
// SPDX-License-Identifier: BSD-2-Clause

package qgo

import (
	"testing"
)

func TestNewNamespace(t *testing.T) {
	tests := []struct {
		name    string
		parts   []string
		wantLen int
		wantStr string
	}{
		{
			name:    "single part",
			parts:   []string{"example"},
			wantLen: 1,
			wantStr: "example",
		},
		{
			name:    "multiple parts",
			parts:   []string{"example", "room", "123"},
			wantLen: 3,
			wantStr: "example/room/123",
		},
		{
			name:    "empty parts",
			parts:   []string{},
			wantLen: 0,
			wantStr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ns := NewNamespace(tt.parts...)
			if ns.NumEntries() != tt.wantLen {
				t.Errorf("NumEntries() = %d, want %d", ns.NumEntries(), tt.wantLen)
			}
			if ns.String() != tt.wantStr {
				t.Errorf("String() = %q, want %q", ns.String(), tt.wantStr)
			}
		})
	}
}

func TestParseNamespace(t *testing.T) {
	tests := []struct {
		input   string
		wantLen int
		wantStr string
	}{
		{"example/room/123", 3, "example/room/123"},
		{"/example/room/", 2, "example/room"},
		{"single", 1, "single"},
		{"", 0, ""},
		{"/", 0, ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			ns := ParseNamespace(tt.input)
			if ns.NumEntries() != tt.wantLen {
				t.Errorf("NumEntries() = %d, want %d", ns.NumEntries(), tt.wantLen)
			}
			if ns.String() != tt.wantStr {
				t.Errorf("String() = %q, want %q", ns.String(), tt.wantStr)
			}
		})
	}
}

func TestNamespace_IsEmpty(t *testing.T) {
	empty := NewNamespace()
	if !empty.IsEmpty() {
		t.Error("Empty namespace should return true for IsEmpty()")
	}

	notEmpty := NewNamespace("test")
	if notEmpty.IsEmpty() {
		t.Error("Non-empty namespace should return false for IsEmpty()")
	}
}

func TestNewTrackName(t *testing.T) {
	name := NewTrackName("video")
	if name.String() != "video" {
		t.Errorf("String() = %q, want %q", name.String(), "video")
	}
	if name.Len() != 5 {
		t.Errorf("Len() = %d, want %d", name.Len(), 5)
	}
}

func TestNewTrackNameFromBytes(t *testing.T) {
	data := []byte("audio")
	name := NewTrackNameFromBytes(data)

	if name.String() != "audio" {
		t.Errorf("String() = %q, want %q", name.String(), "audio")
	}

	// Verify it's a copy
	data[0] = 'x'
	if name.String() != "audio" {
		t.Error("TrackName should copy data, not reference it")
	}
}

func TestFullTrackName_String(t *testing.T) {
	tests := []struct {
		name      string
		namespace Namespace
		trackName TrackName
		want      string
	}{
		{
			name:      "with namespace",
			namespace: NewNamespace("example", "room"),
			trackName: NewTrackName("video"),
			want:      "example/room/video",
		},
		{
			name:      "without namespace",
			namespace: NewNamespace(),
			trackName: NewTrackName("video"),
			want:      "video",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ftn := FullTrackName{
				Namespace: tt.namespace,
				TrackName: tt.trackName,
			}
			if got := ftn.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestObjectStatus_String(t *testing.T) {
	tests := []struct {
		status ObjectStatus
		want   string
	}{
		{ObjectStatusAvailable, "Available"},
		{ObjectStatusDoesNotExist, "DoesNotExist"},
		{ObjectStatusEndOfGroup, "EndOfGroup"},
		{ObjectStatusEndOfTrack, "EndOfTrack"},
		{ObjectStatusEndOfSubGroup, "EndOfSubGroup"},
		{ObjectStatus(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.status.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTrackMode_String(t *testing.T) {
	if TrackModeDatagram.String() != "Datagram" {
		t.Error("TrackModeDatagram.String() should return 'Datagram'")
	}
	if TrackModeStream.String() != "Stream" {
		t.Error("TrackModeStream.String() should return 'Stream'")
	}
}

func TestGroupOrder_String(t *testing.T) {
	tests := []struct {
		order GroupOrder
		want  string
	}{
		{GroupOrderOriginal, "Original"},
		{GroupOrderAscending, "Ascending"},
		{GroupOrderDescending, "Descending"},
	}

	for _, tt := range tests {
		if got := tt.order.String(); got != tt.want {
			t.Errorf("%v.String() = %q, want %q", tt.order, got, tt.want)
		}
	}
}

func TestFilterType_String(t *testing.T) {
	tests := []struct {
		filter FilterType
		want   string
	}{
		{FilterTypeLargestObject, "LargestObject"},
		{FilterTypeLatestGroup, "LatestGroup"},
		{FilterTypeLatestObject, "LatestObject"},
		{FilterTypeAbsoluteStart, "AbsoluteStart"},
		{FilterTypeAbsoluteRange, "AbsoluteRange"},
	}

	for _, tt := range tests {
		if got := tt.filter.String(); got != tt.want {
			t.Errorf("%v.String() = %q, want %q", tt.filter, got, tt.want)
		}
	}
}

func BenchmarkNewNamespace(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewNamespace("example", "room", "123")
	}
}

func BenchmarkParseNamespace(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = ParseNamespace("example/room/123")
	}
}

func BenchmarkNamespace_String(b *testing.B) {
	ns := NewNamespace("example", "room", "123")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ns.String()
	}
}

func BenchmarkNewTrackName(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewTrackName("video-track-001")
	}
}

func BenchmarkFullTrackName_String(b *testing.B) {
	ftn := FullTrackName{
		Namespace: NewNamespace("example", "room", "123"),
		TrackName: NewTrackName("video"),
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ftn.String()
	}
}
