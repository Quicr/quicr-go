// SPDX-FileCopyrightText: Copyright (c) 2025 Cisco Systems
// SPDX-License-Identifier: BSD-2-Clause

package qgo

import (
	"errors"
	"testing"
)

func TestSentinelErrors(t *testing.T) {
	// Verify sentinel errors are distinct
	sentinels := []error{
		ErrClosed,
		ErrNotConnected,
		ErrNotReady,
		ErrInvalidParam,
		ErrTimeout,
		ErrNotAuthorized,
		ErrAlreadyExists,
		ErrNotFound,
		ErrCannotPublish,
		ErrInternal,
	}

	for i, e1 := range sentinels {
		for j, e2 := range sentinels {
			if i != j && errors.Is(e1, e2) {
				t.Errorf("errors %v and %v should be distinct", e1, e2)
			}
		}
	}
}

func TestClientStatus_String(t *testing.T) {
	tests := []struct {
		status ClientStatus
		want   string
	}{
		{ClientStatusReady, "Ready"},
		{ClientStatusNotReady, "NotReady"},
		{ClientStatusInternalError, "InternalError"},
		{ClientStatusInvalidParams, "InvalidParams"},
		{ClientStatusConnecting, "Connecting"},
		{ClientStatusDisconnecting, "Disconnecting"},
		{ClientStatusNotConnected, "NotConnected"},
		{ClientStatusFailedToConnect, "FailedToConnect"},
		{ClientStatusPendingServerSetup, "PendingServerSetup"},
		{ClientStatus(99), "Unknown(99)"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.status.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestClientStatus_IsConnected(t *testing.T) {
	if !ClientStatusReady.IsConnected() {
		t.Error("ClientStatusReady should be connected")
	}

	notConnected := []ClientStatus{
		ClientStatusNotReady,
		ClientStatusInternalError,
		ClientStatusConnecting,
		ClientStatusNotConnected,
		ClientStatusFailedToConnect,
	}

	for _, s := range notConnected {
		if s.IsConnected() {
			t.Errorf("%s should not be connected", s)
		}
	}
}

func TestPublishStatus_String(t *testing.T) {
	tests := []struct {
		status PublishStatus
		want   string
	}{
		{PublishStatusOK, "OK"},
		{PublishStatusNotConnected, "NotConnected"},
		{PublishStatusNotAnnounced, "NotAnnounced"},
		{PublishStatusPendingAnnounceResponse, "PendingAnnounceResponse"},
		{PublishStatusAnnounceNotAuthorized, "AnnounceNotAuthorized"},
		{PublishStatusNoSubscribers, "NoSubscribers"},
		{PublishStatusSendingUnannounce, "SendingUnannounce"},
		{PublishStatusSubscriptionUpdated, "SubscriptionUpdated"},
		{PublishStatusNewGroupRequested, "NewGroupRequested"},
		{PublishStatusPendingPublishOK, "PendingPublishOK"},
		{PublishStatusPaused, "Paused"},
		{PublishStatus(99), "Unknown(99)"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.status.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPublishStatus_CanPublish(t *testing.T) {
	canPublish := []PublishStatus{
		PublishStatusOK,
		PublishStatusNewGroupRequested,
		PublishStatusSubscriptionUpdated,
	}

	for _, s := range canPublish {
		if !s.CanPublish() {
			t.Errorf("%s should allow publishing", s)
		}
	}

	cannotPublish := []PublishStatus{
		PublishStatusNotConnected,
		PublishStatusNotAnnounced,
		PublishStatusPendingAnnounceResponse,
		PublishStatusAnnounceNotAuthorized,
		PublishStatusNoSubscribers,
		PublishStatusPaused,
	}

	for _, s := range cannotPublish {
		if s.CanPublish() {
			t.Errorf("%s should not allow publishing", s)
		}
	}
}

func TestSubscribeStatus_String(t *testing.T) {
	tests := []struct {
		status SubscribeStatus
		want   string
	}{
		{SubscribeStatusOK, "OK"},
		{SubscribeStatusNotConnected, "NotConnected"},
		{SubscribeStatusError, "Error"},
		{SubscribeStatusNotAuthorized, "NotAuthorized"},
		{SubscribeStatusNotSubscribed, "NotSubscribed"},
		{SubscribeStatusPendingResponse, "PendingResponse"},
		{SubscribeStatusSendingUnsubscribe, "SendingUnsubscribe"},
		{SubscribeStatusPaused, "Paused"},
		{SubscribeStatusNewGroupRequested, "NewGroupRequested"},
		{SubscribeStatusCancelled, "Cancelled"},
		{SubscribeStatusDoneByFin, "DoneByFin"},
		{SubscribeStatusDoneByReset, "DoneByReset"},
		{SubscribeStatus(99), "Unknown(99)"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.status.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSubscribeStatus_IsActive(t *testing.T) {
	if !SubscribeStatusOK.IsActive() {
		t.Error("SubscribeStatusOK should be active")
	}

	notActive := []SubscribeStatus{
		SubscribeStatusNotConnected,
		SubscribeStatusError,
		SubscribeStatusNotAuthorized,
		SubscribeStatusNotSubscribed,
		SubscribeStatusPaused,
		SubscribeStatusCancelled,
	}

	for _, s := range notActive {
		if s.IsActive() {
			t.Errorf("%s should not be active", s)
		}
	}
}

func TestPublishObjectStatus_String(t *testing.T) {
	tests := []struct {
		status PublishObjectStatus
		want   string
	}{
		{PublishObjectOK, "OK"},
		{PublishObjectInternalError, "InternalError"},
		{PublishObjectNotAuthorized, "NotAuthorized"},
		{PublishObjectNotAnnounced, "NotAnnounced"},
		{PublishObjectNoSubscribers, "NoSubscribers"},
		{PublishObjectPayloadLengthExceeded, "PayloadLengthExceeded"},
		{PublishObjectPreviousTruncated, "PreviousTruncated"},
		{PublishObjectNoPrevious, "NoPrevious"},
		{PublishObjectDataComplete, "DataComplete"},
		{PublishObjectContinuationNeeded, "ContinuationNeeded"},
		{PublishObjectDataIncomplete, "DataIncomplete"},
		{PublishObjectDataTooLarge, "DataTooLarge"},
		{PublishObjectMustStartNewGroup, "MustStartNewGroup"},
		{PublishObjectMustStartNewTrack, "MustStartNewTrack"},
		{PublishObjectPaused, "Paused"},
		{PublishObjectPendingOK, "PendingOK"},
		{PublishObjectStatus(99), "Unknown(99)"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.status.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPublishObjectStatus_IsSuccess(t *testing.T) {
	successful := []PublishObjectStatus{
		PublishObjectOK,
		PublishObjectDataComplete,
	}

	for _, s := range successful {
		if !s.IsSuccess() {
			t.Errorf("%s should be success", s)
		}
	}

	notSuccessful := []PublishObjectStatus{
		PublishObjectInternalError,
		PublishObjectNotAuthorized,
		PublishObjectNotAnnounced,
		PublishObjectPayloadLengthExceeded,
	}

	for _, s := range notSuccessful {
		if s.IsSuccess() {
			t.Errorf("%s should not be success", s)
		}
	}
}

func TestPublishObjectStatus_Error(t *testing.T) {
	// Successful statuses should return nil error
	if err := PublishObjectOK.Error(); err != nil {
		t.Errorf("PublishObjectOK.Error() = %v, want nil", err)
	}
	if err := PublishObjectDataComplete.Error(); err != nil {
		t.Errorf("PublishObjectDataComplete.Error() = %v, want nil", err)
	}

	// NoSubscribers should return nil (not an error)
	if err := PublishObjectNoSubscribers.Error(); err != nil {
		t.Errorf("PublishObjectNoSubscribers.Error() = %v, want nil", err)
	}

	// NotAuthorized should return ErrNotAuthorized
	if err := PublishObjectNotAuthorized.Error(); !errors.Is(err, ErrNotAuthorized) {
		t.Errorf("PublishObjectNotAuthorized.Error() = %v, want ErrNotAuthorized", err)
	}

	// NotAnnounced should return ErrCannotPublish
	if err := PublishObjectNotAnnounced.Error(); !errors.Is(err, ErrCannotPublish) {
		t.Errorf("PublishObjectNotAnnounced.Error() = %v, want ErrCannotPublish", err)
	}

	// Other errors should return a descriptive error
	if err := PublishObjectInternalError.Error(); err == nil {
		t.Error("PublishObjectInternalError.Error() should return an error")
	}
}
