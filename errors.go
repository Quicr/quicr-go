// SPDX-FileCopyrightText: Copyright (c) 2025 Cisco Systems
// SPDX-License-Identifier: BSD-2-Clause

package qgo

import (
	"errors"
	"fmt"
)

// Sentinel errors for common error conditions.
var (
	// ErrClosed indicates the client or handler has been closed.
	ErrClosed = errors.New("qgo: closed")

	// ErrNotConnected indicates the client is not connected.
	ErrNotConnected = errors.New("qgo: not connected")

	// ErrNotReady indicates the client is not ready for operations.
	ErrNotReady = errors.New("qgo: not ready")

	// ErrInvalidParam indicates an invalid parameter was provided.
	ErrInvalidParam = errors.New("qgo: invalid parameter")

	// ErrTimeout indicates an operation timed out.
	ErrTimeout = errors.New("qgo: timeout")

	// ErrNotAuthorized indicates the operation was not authorized.
	ErrNotAuthorized = errors.New("qgo: not authorized")

	// ErrAlreadyExists indicates the resource already exists.
	ErrAlreadyExists = errors.New("qgo: already exists")

	// ErrNotFound indicates the resource was not found.
	ErrNotFound = errors.New("qgo: not found")

	// ErrCannotPublish indicates publishing is not currently allowed.
	ErrCannotPublish = errors.New("qgo: cannot publish")

	// ErrInternal indicates an internal error occurred.
	ErrInternal = errors.New("qgo: internal error")
)

// OpError represents an error that occurred during a specific operation.
type OpError struct {
	Op  string // Operation name (e.g., "connect", "publish", "subscribe")
	Err error  // Underlying error
}

func (e *OpError) Error() string {
	if e.Err == nil {
		return fmt.Sprintf("qgo: %s failed", e.Op)
	}
	return fmt.Sprintf("qgo: %s: %v", e.Op, e.Err)
}

func (e *OpError) Unwrap() error {
	return e.Err
}

// wrapError wraps an error with operation context.
func wrapError(op string, err error) error {
	if err == nil {
		return nil
	}
	return &OpError{Op: op, Err: err}
}

// PublishError represents an error that occurred during a publish operation.
type PublishError struct {
	Op      string              // Operation name
	Status  PublishObjectStatus // The status returned by the publish operation
	Err     error               // Underlying error (may be nil)
}

func (e *PublishError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("qgo: %s: %s: %v", e.Op, e.Status, e.Err)
	}
	return fmt.Sprintf("qgo: %s: %s", e.Op, e.Status)
}

func (e *PublishError) Unwrap() error {
	return e.Err
}

// SubscribeError represents an error that occurred during a subscribe operation.
type SubscribeError struct {
	Op     string          // Operation name
	Status SubscribeStatus // The status at the time of error
	Err    error           // Underlying error (may be nil)
}

func (e *SubscribeError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("qgo: %s: %s: %v", e.Op, e.Status, e.Err)
	}
	return fmt.Sprintf("qgo: %s: %s", e.Op, e.Status)
}

// ClientStatus represents the connection status of a Client.
type ClientStatus uint8

const (
	// ClientStatusReady indicates the client is connected and ready.
	ClientStatusReady ClientStatus = iota
	// ClientStatusNotReady indicates the client is not ready.
	ClientStatusNotReady
	// ClientStatusInternalError indicates an internal error occurred.
	ClientStatusInternalError
	// ClientStatusInvalidParams indicates invalid parameters were provided.
	ClientStatusInvalidParams
	// ClientStatusConnecting indicates the client is connecting.
	ClientStatusConnecting
	// ClientStatusDisconnecting indicates the client is disconnecting.
	ClientStatusDisconnecting
	// ClientStatusNotConnected indicates the client is not connected.
	ClientStatusNotConnected
	// ClientStatusFailedToConnect indicates connection failed.
	ClientStatusFailedToConnect
	// ClientStatusPendingServerSetup indicates waiting for server setup.
	ClientStatusPendingServerSetup
)

// String returns a string representation of the client status.
func (s ClientStatus) String() string {
	switch s {
	case ClientStatusReady:
		return "Ready"
	case ClientStatusNotReady:
		return "NotReady"
	case ClientStatusInternalError:
		return "InternalError"
	case ClientStatusInvalidParams:
		return "InvalidParams"
	case ClientStatusConnecting:
		return "Connecting"
	case ClientStatusDisconnecting:
		return "Disconnecting"
	case ClientStatusNotConnected:
		return "NotConnected"
	case ClientStatusFailedToConnect:
		return "FailedToConnect"
	case ClientStatusPendingServerSetup:
		return "PendingServerSetup"
	default:
		return fmt.Sprintf("Unknown(%d)", s)
	}
}

// IsConnected returns true if the status indicates a connected state.
func (s ClientStatus) IsConnected() bool {
	return s == ClientStatusReady
}

// PublishStatus represents the status of a publish track handler.
type PublishStatus uint8

const (
	// PublishStatusOK indicates the track is ready to publish.
	PublishStatusOK PublishStatus = iota
	// PublishStatusNotConnected indicates the client is not connected.
	PublishStatusNotConnected
	// PublishStatusNotAnnounced indicates the namespace is not announced.
	PublishStatusNotAnnounced
	// PublishStatusPendingAnnounceResponse indicates waiting for announce response.
	PublishStatusPendingAnnounceResponse
	// PublishStatusAnnounceNotAuthorized indicates announce was not authorized.
	PublishStatusAnnounceNotAuthorized
	// PublishStatusNoSubscribers indicates there are no subscribers.
	PublishStatusNoSubscribers
	// PublishStatusSendingUnannounce indicates unannounce is in progress.
	PublishStatusSendingUnannounce
	// PublishStatusSubscriptionUpdated indicates subscription was updated.
	PublishStatusSubscriptionUpdated
	// PublishStatusNewGroupRequested indicates a new group was requested.
	PublishStatusNewGroupRequested
	// PublishStatusPendingPublishOK indicates waiting for publish OK.
	PublishStatusPendingPublishOK
	// PublishStatusPaused indicates publishing is paused.
	PublishStatusPaused
)

// String returns a string representation of the publish status.
func (s PublishStatus) String() string {
	switch s {
	case PublishStatusOK:
		return "OK"
	case PublishStatusNotConnected:
		return "NotConnected"
	case PublishStatusNotAnnounced:
		return "NotAnnounced"
	case PublishStatusPendingAnnounceResponse:
		return "PendingAnnounceResponse"
	case PublishStatusAnnounceNotAuthorized:
		return "AnnounceNotAuthorized"
	case PublishStatusNoSubscribers:
		return "NoSubscribers"
	case PublishStatusSendingUnannounce:
		return "SendingUnannounce"
	case PublishStatusSubscriptionUpdated:
		return "SubscriptionUpdated"
	case PublishStatusNewGroupRequested:
		return "NewGroupRequested"
	case PublishStatusPendingPublishOK:
		return "PendingPublishOK"
	case PublishStatusPaused:
		return "Paused"
	default:
		return fmt.Sprintf("Unknown(%d)", s)
	}
}

// CanPublish returns true if the status allows publishing.
func (s PublishStatus) CanPublish() bool {
	switch s {
	case PublishStatusOK, PublishStatusNewGroupRequested, PublishStatusSubscriptionUpdated:
		return true
	default:
		return false
	}
}

// SubscribeStatus represents the status of a subscribe track handler.
type SubscribeStatus uint8

const (
	// SubscribeStatusOK indicates the subscription is active.
	SubscribeStatusOK SubscribeStatus = iota
	// SubscribeStatusNotConnected indicates the client is not connected.
	SubscribeStatusNotConnected
	// SubscribeStatusError indicates a subscription error occurred.
	SubscribeStatusError
	// SubscribeStatusNotAuthorized indicates the subscription was not authorized.
	SubscribeStatusNotAuthorized
	// SubscribeStatusNotSubscribed indicates not currently subscribed.
	SubscribeStatusNotSubscribed
	// SubscribeStatusPendingResponse indicates waiting for subscribe response.
	SubscribeStatusPendingResponse
	// SubscribeStatusSendingUnsubscribe indicates unsubscribe is in progress.
	SubscribeStatusSendingUnsubscribe
	// SubscribeStatusPaused indicates the subscription is paused.
	SubscribeStatusPaused
	// SubscribeStatusNewGroupRequested indicates a new group was requested.
	SubscribeStatusNewGroupRequested
	// SubscribeStatusCancelled indicates the subscription was cancelled.
	SubscribeStatusCancelled
	// SubscribeStatusDoneByFin indicates the subscription ended with FIN.
	SubscribeStatusDoneByFin
	// SubscribeStatusDoneByReset indicates the subscription ended with RESET.
	SubscribeStatusDoneByReset
)

// String returns a string representation of the subscribe status.
func (s SubscribeStatus) String() string {
	switch s {
	case SubscribeStatusOK:
		return "OK"
	case SubscribeStatusNotConnected:
		return "NotConnected"
	case SubscribeStatusError:
		return "Error"
	case SubscribeStatusNotAuthorized:
		return "NotAuthorized"
	case SubscribeStatusNotSubscribed:
		return "NotSubscribed"
	case SubscribeStatusPendingResponse:
		return "PendingResponse"
	case SubscribeStatusSendingUnsubscribe:
		return "SendingUnsubscribe"
	case SubscribeStatusPaused:
		return "Paused"
	case SubscribeStatusNewGroupRequested:
		return "NewGroupRequested"
	case SubscribeStatusCancelled:
		return "Cancelled"
	case SubscribeStatusDoneByFin:
		return "DoneByFin"
	case SubscribeStatusDoneByReset:
		return "DoneByReset"
	default:
		return fmt.Sprintf("Unknown(%d)", s)
	}
}

// IsActive returns true if the subscription is currently receiving data.
func (s SubscribeStatus) IsActive() bool {
	return s == SubscribeStatusOK
}

// PublishObjectStatus represents the result of a publish operation.
type PublishObjectStatus uint8

const (
	// PublishObjectOK indicates the object was published successfully.
	PublishObjectOK PublishObjectStatus = iota
	// PublishObjectInternalError indicates an internal error.
	PublishObjectInternalError
	// PublishObjectNotAuthorized indicates not authorized to publish.
	PublishObjectNotAuthorized
	// PublishObjectNotAnnounced indicates namespace not announced.
	PublishObjectNotAnnounced
	// PublishObjectNoSubscribers indicates no subscribers.
	PublishObjectNoSubscribers
	// PublishObjectPayloadLengthExceeded indicates payload too large.
	PublishObjectPayloadLengthExceeded
	// PublishObjectPreviousTruncated indicates previous object was truncated.
	PublishObjectPreviousTruncated
	// PublishObjectNoPrevious indicates no previous object exists.
	PublishObjectNoPrevious
	// PublishObjectDataComplete indicates object data is complete.
	PublishObjectDataComplete
	// PublishObjectContinuationNeeded indicates more data is needed.
	PublishObjectContinuationNeeded
	// PublishObjectDataIncomplete indicates object data is incomplete.
	PublishObjectDataIncomplete
	// PublishObjectDataTooLarge indicates data exceeds payload length.
	PublishObjectDataTooLarge
	// PublishObjectMustStartNewGroup indicates a new group must be started.
	PublishObjectMustStartNewGroup
	// PublishObjectMustStartNewTrack indicates a new track must be started.
	PublishObjectMustStartNewTrack
	// PublishObjectPaused indicates publishing is paused.
	PublishObjectPaused
	// PublishObjectPendingOK indicates waiting for publish OK.
	PublishObjectPendingOK
)

// String returns a string representation of the publish object status.
func (s PublishObjectStatus) String() string {
	switch s {
	case PublishObjectOK:
		return "OK"
	case PublishObjectInternalError:
		return "InternalError"
	case PublishObjectNotAuthorized:
		return "NotAuthorized"
	case PublishObjectNotAnnounced:
		return "NotAnnounced"
	case PublishObjectNoSubscribers:
		return "NoSubscribers"
	case PublishObjectPayloadLengthExceeded:
		return "PayloadLengthExceeded"
	case PublishObjectPreviousTruncated:
		return "PreviousTruncated"
	case PublishObjectNoPrevious:
		return "NoPrevious"
	case PublishObjectDataComplete:
		return "DataComplete"
	case PublishObjectContinuationNeeded:
		return "ContinuationNeeded"
	case PublishObjectDataIncomplete:
		return "DataIncomplete"
	case PublishObjectDataTooLarge:
		return "DataTooLarge"
	case PublishObjectMustStartNewGroup:
		return "MustStartNewGroup"
	case PublishObjectMustStartNewTrack:
		return "MustStartNewTrack"
	case PublishObjectPaused:
		return "Paused"
	case PublishObjectPendingOK:
		return "PendingOK"
	default:
		return fmt.Sprintf("Unknown(%d)", s)
	}
}

// IsSuccess returns true if the publish was successful.
func (s PublishObjectStatus) IsSuccess() bool {
	return s == PublishObjectOK || s == PublishObjectDataComplete
}

// Error returns an error if the publish failed, nil otherwise.
func (s PublishObjectStatus) Error() error {
	switch s {
	case PublishObjectOK, PublishObjectDataComplete:
		return nil
	case PublishObjectNotAuthorized:
		return ErrNotAuthorized
	case PublishObjectNotAnnounced:
		return ErrCannotPublish
	case PublishObjectNoSubscribers:
		return nil // Not an error, just no one listening
	default:
		return fmt.Errorf("qgo: publish failed: %s", s.String())
	}
}
