// SPDX-FileCopyrightText: Copyright (c) 2025 Cisco Systems
// SPDX-License-Identifier: BSD-2-Clause

package qgo

import (
	"errors"
	"testing"
	"time"
)

func TestClientConfig_Validate(t *testing.T) {
	t.Run("empty URI", func(t *testing.T) {
		cfg := ClientConfig{}
		err := cfg.Validate()
		if !errors.Is(err, ErrInvalidParam) {
			t.Errorf("Validate() = %v, want ErrInvalidParam", err)
		}
	})

	t.Run("valid config with defaults", func(t *testing.T) {
		cfg := ClientConfig{
			ConnectURI: "moqt://localhost:4433",
		}
		err := cfg.Validate()
		if err != nil {
			t.Errorf("Validate() = %v, want nil", err)
		}

		// Check defaults were applied
		if cfg.EndpointID != "qgo-client" {
			t.Errorf("EndpointID = %q, want %q", cfg.EndpointID, "qgo-client")
		}
		if cfg.MetricsSampleInterval != DefaultMetricsSampleInterval {
			t.Errorf("MetricsSampleInterval = %v, want %v", cfg.MetricsSampleInterval, DefaultMetricsSampleInterval)
		}
		if cfg.TickServiceDelay != DefaultTickServiceDelay {
			t.Errorf("TickServiceDelay = %v, want %v", cfg.TickServiceDelay, DefaultTickServiceDelay)
		}
	})

	t.Run("custom values preserved", func(t *testing.T) {
		cfg := ClientConfig{
			ConnectURI:            "moqt://relay.example.com:4433",
			EndpointID:            "custom-client",
			MetricsSampleInterval: 10 * time.Second,
			TickServiceDelay:      500 * time.Microsecond,
		}
		err := cfg.Validate()
		if err != nil {
			t.Errorf("Validate() = %v, want nil", err)
		}

		// Check custom values were preserved
		if cfg.EndpointID != "custom-client" {
			t.Errorf("EndpointID = %q, want %q", cfg.EndpointID, "custom-client")
		}
		if cfg.MetricsSampleInterval != 10*time.Second {
			t.Errorf("MetricsSampleInterval = %v, want %v", cfg.MetricsSampleInterval, 10*time.Second)
		}
		if cfg.TickServiceDelay != 500*time.Microsecond {
			t.Errorf("TickServiceDelay = %v, want %v", cfg.TickServiceDelay, 500*time.Microsecond)
		}
	})
}

func TestPublishTrackConfig_Validate(t *testing.T) {
	t.Run("empty namespace", func(t *testing.T) {
		cfg := PublishTrackConfig{}
		err := cfg.Validate()
		if !errors.Is(err, ErrInvalidParam) {
			t.Errorf("Validate() = %v, want ErrInvalidParam", err)
		}
	})

	t.Run("valid config with defaults", func(t *testing.T) {
		cfg := PublishTrackConfig{
			FullTrackName: FullTrackName{
				Namespace: NewNamespace("example", "room"),
				TrackName: NewTrackName("video"),
			},
		}
		err := cfg.Validate()
		if err != nil {
			t.Errorf("Validate() = %v, want nil", err)
		}

		// Check defaults were applied
		if cfg.Priority != 128 {
			t.Errorf("Priority = %d, want %d", cfg.Priority, 128)
		}
		if cfg.TTL != 5000 {
			t.Errorf("TTL = %d, want %d", cfg.TTL, 5000)
		}
	})

	t.Run("custom values preserved", func(t *testing.T) {
		cfg := PublishTrackConfig{
			FullTrackName: FullTrackName{
				Namespace: NewNamespace("example"),
				TrackName: NewTrackName("audio"),
			},
			TrackMode:   TrackModeDatagram,
			Priority:    64,
			TTL:         10000,
			UseAnnounce: true,
		}
		err := cfg.Validate()
		if err != nil {
			t.Errorf("Validate() = %v, want nil", err)
		}

		// Check custom values were preserved
		if cfg.Priority != 64 {
			t.Errorf("Priority = %d, want %d", cfg.Priority, 64)
		}
		if cfg.TTL != 10000 {
			t.Errorf("TTL = %d, want %d", cfg.TTL, 10000)
		}
		if cfg.TrackMode != TrackModeDatagram {
			t.Errorf("TrackMode = %v, want %v", cfg.TrackMode, TrackModeDatagram)
		}
		if !cfg.UseAnnounce {
			t.Error("UseAnnounce should be true")
		}
	})
}

func TestSubscribeTrackConfig_Validate(t *testing.T) {
	t.Run("empty namespace", func(t *testing.T) {
		cfg := SubscribeTrackConfig{}
		err := cfg.Validate()
		if !errors.Is(err, ErrInvalidParam) {
			t.Errorf("Validate() = %v, want ErrInvalidParam", err)
		}
	})

	t.Run("valid config with defaults", func(t *testing.T) {
		cfg := SubscribeTrackConfig{
			FullTrackName: FullTrackName{
				Namespace: NewNamespace("example", "room"),
				TrackName: NewTrackName("video"),
			},
		}
		err := cfg.Validate()
		if err != nil {
			t.Errorf("Validate() = %v, want nil", err)
		}

		// Check defaults were applied
		if cfg.Priority != 128 {
			t.Errorf("Priority = %d, want %d", cfg.Priority, 128)
		}
	})

	t.Run("custom values preserved", func(t *testing.T) {
		cfg := SubscribeTrackConfig{
			FullTrackName: FullTrackName{
				Namespace: NewNamespace("example"),
				TrackName: NewTrackName("audio"),
			},
			Priority:   64,
			GroupOrder: GroupOrderDescending,
			FilterType: FilterTypeLatestGroup,
		}
		err := cfg.Validate()
		if err != nil {
			t.Errorf("Validate() = %v, want nil", err)
		}

		// Check custom values were preserved
		if cfg.Priority != 64 {
			t.Errorf("Priority = %d, want %d", cfg.Priority, 64)
		}
		if cfg.GroupOrder != GroupOrderDescending {
			t.Errorf("GroupOrder = %v, want %v", cfg.GroupOrder, GroupOrderDescending)
		}
		if cfg.FilterType != FilterTypeLatestGroup {
			t.Errorf("FilterType = %v, want %v", cfg.FilterType, FilterTypeLatestGroup)
		}
	})
}

func TestDefaultConstants(t *testing.T) {
	if DefaultMetricsSampleInterval != 5*time.Second {
		t.Errorf("DefaultMetricsSampleInterval = %v, want %v", DefaultMetricsSampleInterval, 5*time.Second)
	}
	if DefaultTickServiceDelay != 333*time.Microsecond {
		t.Errorf("DefaultTickServiceDelay = %v, want %v", DefaultTickServiceDelay, 333*time.Microsecond)
	}
}
