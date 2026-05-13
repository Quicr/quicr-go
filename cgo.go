// SPDX-FileCopyrightText: Copyright (c) 2025 Cisco Systems
// SPDX-License-Identifier: BSD-2-Clause

package qgo

/*
#cgo CFLAGS: -I${SRCDIR}/cshim -I${SRCDIR}/libquicr/include
#cgo LDFLAGS: -L${SRCDIR}/build -L${SRCDIR}/build/lib -L${SRCDIR}/build/libquicr/src -L${SRCDIR}/build/libquicr/dependencies/spdlog -L${SRCDIR}/build/libquicr/dependencies/picoquic -L${SRCDIR}/build/_deps/picotls-build -lquicr_shim -lquicr -lpicoquic-core -lpicohttp-core -lpicoquic-log -lpicotls-openssl -lpicotls-core -lpicotls-minicrypto -lspdlog -lstdc++ -lm -lpthread
#cgo darwin LDFLAGS: -L/opt/homebrew/opt/openssl@3/lib -lssl -lcrypto -framework CoreFoundation -framework Security -lresolv
#cgo linux LDFLAGS: -lssl -lcrypto

#include "quicr_shim.h"
#include <stdlib.h>
#include <string.h>
*/
import "C"

import (
	"unsafe"
)

// C type aliases for cleaner Go code
type (
	cClient                     = C.quicr_client_t
	cPublishTrackHandler        = C.quicr_publish_track_handler_t
	cSubscribeTrackHandler      = C.quicr_subscribe_track_handler_t
	cPublishNamespaceHandler    = C.quicr_publish_namespace_handler_t
	cSubscribeNamespaceHandler  = C.quicr_subscribe_namespace_handler_t
	cResult                     = C.quicr_result_t
	cClientStatus               = C.quicr_client_status_t
	cPublishStatus              = C.quicr_publish_status_t
	cSubscribeStatus            = C.quicr_subscribe_status_t
	cPublishNamespaceStatus     = C.quicr_publish_namespace_status_t
	cSubscribeNamespaceStatus   = C.quicr_subscribe_namespace_status_t
	cPublishObjectStatus        = C.quicr_publish_object_status_t
	cClientConfig               = C.quicr_client_config_t
	cPublishTrackConfig         = C.quicr_publish_track_config_t
	cSubscribeTrackConfig       = C.quicr_subscribe_track_config_t
	cObjectHeaders              = C.quicr_object_headers_t
	cObject                     = C.quicr_object_t
	cNamespace                  = C.quicr_namespace_t
	cNamespaceEntry             = C.quicr_namespace_entry_t
	cTrackName                  = C.quicr_track_name_t
	cFullTrackName              = C.quicr_full_track_name_t
)

// resultToError converts a C result to a Go error.
func resultToError(result cResult) error {
	switch result {
	case C.QUICR_OK:
		return nil
	case C.QUICR_ERROR_INVALID_PARAM:
		return ErrInvalidParam
	case C.QUICR_ERROR_NOT_CONNECTED:
		return ErrNotConnected
	case C.QUICR_ERROR_NOT_READY:
		return ErrNotReady
	case C.QUICR_ERROR_TIMEOUT:
		return ErrTimeout
	case C.QUICR_ERROR_NOT_AUTHORIZED:
		return ErrNotAuthorized
	case C.QUICR_ERROR_ALREADY_EXISTS:
		return ErrAlreadyExists
	case C.QUICR_ERROR_NOT_FOUND:
		return ErrNotFound
	default:
		return ErrInternal
	}
}

// convertClientStatus converts C client status to Go.
func convertClientStatus(status cClientStatus) ClientStatus {
	switch status {
	case C.QUICR_CLIENT_STATUS_READY:
		return ClientStatusReady
	case C.QUICR_CLIENT_STATUS_NOT_READY:
		return ClientStatusNotReady
	case C.QUICR_CLIENT_STATUS_INTERNAL_ERROR:
		return ClientStatusInternalError
	case C.QUICR_CLIENT_STATUS_INVALID_PARAMS:
		return ClientStatusInvalidParams
	case C.QUICR_CLIENT_STATUS_CONNECTING:
		return ClientStatusConnecting
	case C.QUICR_CLIENT_STATUS_DISCONNECTING:
		return ClientStatusDisconnecting
	case C.QUICR_CLIENT_STATUS_NOT_CONNECTED:
		return ClientStatusNotConnected
	case C.QUICR_CLIENT_STATUS_FAILED_TO_CONNECT:
		return ClientStatusFailedToConnect
	case C.QUICR_CLIENT_STATUS_PENDING_SERVER_SETUP:
		return ClientStatusPendingServerSetup
	default:
		return ClientStatusInternalError
	}
}

// convertPublishStatus converts C publish status to Go.
func convertPublishStatus(status cPublishStatus) PublishStatus {
	switch status {
	case C.QUICR_PUBLISH_STATUS_OK:
		return PublishStatusOK
	case C.QUICR_PUBLISH_STATUS_NOT_CONNECTED:
		return PublishStatusNotConnected
	case C.QUICR_PUBLISH_STATUS_NOT_ANNOUNCED:
		return PublishStatusNotAnnounced
	case C.QUICR_PUBLISH_STATUS_PENDING_ANNOUNCE_RESPONSE:
		return PublishStatusPendingAnnounceResponse
	case C.QUICR_PUBLISH_STATUS_ANNOUNCE_NOT_AUTHORIZED:
		return PublishStatusAnnounceNotAuthorized
	case C.QUICR_PUBLISH_STATUS_NO_SUBSCRIBERS:
		return PublishStatusNoSubscribers
	case C.QUICR_PUBLISH_STATUS_SENDING_UNANNOUNCE:
		return PublishStatusSendingUnannounce
	case C.QUICR_PUBLISH_STATUS_SUBSCRIPTION_UPDATED:
		return PublishStatusSubscriptionUpdated
	case C.QUICR_PUBLISH_STATUS_NEW_GROUP_REQUESTED:
		return PublishStatusNewGroupRequested
	case C.QUICR_PUBLISH_STATUS_PENDING_PUBLISH_OK:
		return PublishStatusPendingPublishOK
	case C.QUICR_PUBLISH_STATUS_PAUSED:
		return PublishStatusPaused
	default:
		return PublishStatusNotConnected
	}
}

// convertSubscribeStatus converts C subscribe status to Go.
func convertSubscribeStatus(status cSubscribeStatus) SubscribeStatus {
	switch status {
	case C.QUICR_SUBSCRIBE_STATUS_OK:
		return SubscribeStatusOK
	case C.QUICR_SUBSCRIBE_STATUS_NOT_CONNECTED:
		return SubscribeStatusNotConnected
	case C.QUICR_SUBSCRIBE_STATUS_ERROR:
		return SubscribeStatusError
	case C.QUICR_SUBSCRIBE_STATUS_NOT_AUTHORIZED:
		return SubscribeStatusNotAuthorized
	case C.QUICR_SUBSCRIBE_STATUS_NOT_SUBSCRIBED:
		return SubscribeStatusNotSubscribed
	case C.QUICR_SUBSCRIBE_STATUS_PENDING_RESPONSE:
		return SubscribeStatusPendingResponse
	case C.QUICR_SUBSCRIBE_STATUS_SENDING_UNSUBSCRIBE:
		return SubscribeStatusSendingUnsubscribe
	case C.QUICR_SUBSCRIBE_STATUS_PAUSED:
		return SubscribeStatusPaused
	case C.QUICR_SUBSCRIBE_STATUS_NEW_GROUP_REQUESTED:
		return SubscribeStatusNewGroupRequested
	case C.QUICR_SUBSCRIBE_STATUS_CANCELLED:
		return SubscribeStatusCancelled
	case C.QUICR_SUBSCRIBE_STATUS_DONE_BY_FIN:
		return SubscribeStatusDoneByFin
	case C.QUICR_SUBSCRIBE_STATUS_DONE_BY_RESET:
		return SubscribeStatusDoneByReset
	default:
		return SubscribeStatusError
	}
}

// convertPublishObjectStatus converts C publish object status to Go.
func convertPublishObjectStatus(status cPublishObjectStatus) PublishObjectStatus {
	switch status {
	case C.QUICR_PUBLISH_OBJECT_OK:
		return PublishObjectOK
	case C.QUICR_PUBLISH_OBJECT_INTERNAL_ERROR:
		return PublishObjectInternalError
	case C.QUICR_PUBLISH_OBJECT_NOT_AUTHORIZED:
		return PublishObjectNotAuthorized
	case C.QUICR_PUBLISH_OBJECT_NOT_ANNOUNCED:
		return PublishObjectNotAnnounced
	case C.QUICR_PUBLISH_OBJECT_NO_SUBSCRIBERS:
		return PublishObjectNoSubscribers
	case C.QUICR_PUBLISH_OBJECT_PAYLOAD_LENGTH_EXCEEDED:
		return PublishObjectPayloadLengthExceeded
	case C.QUICR_PUBLISH_OBJECT_PREVIOUS_TRUNCATED:
		return PublishObjectPreviousTruncated
	case C.QUICR_PUBLISH_OBJECT_NO_PREVIOUS:
		return PublishObjectNoPrevious
	case C.QUICR_PUBLISH_OBJECT_DATA_COMPLETE:
		return PublishObjectDataComplete
	case C.QUICR_PUBLISH_OBJECT_CONTINUATION_NEEDED:
		return PublishObjectContinuationNeeded
	case C.QUICR_PUBLISH_OBJECT_DATA_INCOMPLETE:
		return PublishObjectDataIncomplete
	case C.QUICR_PUBLISH_OBJECT_DATA_TOO_LARGE:
		return PublishObjectDataTooLarge
	case C.QUICR_PUBLISH_OBJECT_MUST_START_NEW_GROUP:
		return PublishObjectMustStartNewGroup
	case C.QUICR_PUBLISH_OBJECT_MUST_START_NEW_TRACK:
		return PublishObjectMustStartNewTrack
	case C.QUICR_PUBLISH_OBJECT_PAUSED:
		return PublishObjectPaused
	case C.QUICR_PUBLISH_OBJECT_PENDING_OK:
		return PublishObjectPendingOK
	default:
		return PublishObjectInternalError
	}
}

// namespaceToC converts a Go Namespace to C.
func namespaceToC(ns Namespace) cNamespace {
	var cNs cNamespace
	cNs.num_entries = C.uint8_t(len(ns.entries))

	for i, entry := range ns.entries {
		if i >= C.QUICR_MAX_NAMESPACE_ENTRIES {
			break
		}
		entryLen := len(entry)
		if entryLen > C.QUICR_MAX_ENTRY_SIZE {
			entryLen = C.QUICR_MAX_ENTRY_SIZE
		}
		cNs.entries[i].len = C.uint16_t(entryLen)
		if entryLen > 0 {
			C.memcpy(
				unsafe.Pointer(&cNs.entries[i].data[0]),
				unsafe.Pointer(&entry[0]),
				C.size_t(entryLen),
			)
		}
	}

	return cNs
}

// trackNameToC converts a Go TrackName to C.
func trackNameToC(tn TrackName) cTrackName {
	var cTn cTrackName
	nameLen := len(tn.data)
	if nameLen > C.QUICR_MAX_TRACK_NAME_SIZE {
		nameLen = C.QUICR_MAX_TRACK_NAME_SIZE
	}
	cTn.len = C.uint16_t(nameLen)
	if nameLen > 0 {
		C.memcpy(
			unsafe.Pointer(&cTn.data[0]),
			unsafe.Pointer(&tn.data[0]),
			C.size_t(nameLen),
		)
	}
	return cTn
}

// fullTrackNameToC converts a Go FullTrackName to C.
func fullTrackNameToC(ftn FullTrackName) cFullTrackName {
	var cFtn cFullTrackName
	cFtn.name_space = namespaceToC(ftn.Namespace)
	cFtn.name = trackNameToC(ftn.TrackName)
	return cFtn
}

// objectHeadersToC converts Go ObjectHeaders to C.
func objectHeadersToC(h ObjectHeaders) cObjectHeaders {
	var cH cObjectHeaders
	cH.group_id = C.uint64_t(h.GroupID)
	cH.object_id = C.uint64_t(h.ObjectID)
	cH.subgroup_id = C.uint64_t(h.SubgroupID)
	cH.status = C.quicr_object_status_t(h.Status)
	cH.priority = C.uint8_t(h.Priority)
	cH.ttl = C.uint16_t(h.TTL)
	cH.has_priority = 1
	cH.has_ttl = 1
	cH.has_track_mode = 0
	return cH
}

// convertObject converts a C object to Go.
func convertObject(cObj *cObject) Object {
	obj := Object{
		Headers: ObjectHeaders{
			GroupID:    uint64(cObj.headers.group_id),
			ObjectID:   uint64(cObj.headers.object_id),
			SubgroupID: uint64(cObj.headers.subgroup_id),
			Priority:   uint8(cObj.headers.priority),
			TTL:        uint16(cObj.headers.ttl),
			Status:     ObjectStatus(cObj.headers.status),
		},
	}

	// Copy data to Go memory (critical: C data may be freed after callback returns)
	if cObj.data_len > 0 && cObj.data != nil {
		obj.Data = C.GoBytes(unsafe.Pointer(cObj.data), C.int(cObj.data_len))
	}

	return obj
}

// convertNamespace converts a C namespace to Go.
func convertNamespace(cNs *cNamespace) Namespace {
	numEntries := int(cNs.num_entries)
	entries := make([][]byte, numEntries)
	for i := 0; i < numEntries; i++ {
		entryLen := int(cNs.entries[i].len)
		if entryLen > 0 {
			entries[i] = C.GoBytes(unsafe.Pointer(&cNs.entries[i].data[0]), C.int(entryLen))
		} else {
			entries[i] = []byte{}
		}
	}
	return Namespace{entries: entries}
}

// convertFullTrackName converts a C full track name to Go.
func convertFullTrackName(cFtn *cFullTrackName) FullTrackName {
	// Convert namespace
	numEntries := int(cFtn.name_space.num_entries)
	entries := make([][]byte, numEntries)
	for i := 0; i < numEntries; i++ {
		entryLen := int(cFtn.name_space.entries[i].len)
		if entryLen > 0 {
			entries[i] = C.GoBytes(unsafe.Pointer(&cFtn.name_space.entries[i].data[0]), C.int(entryLen))
		} else {
			entries[i] = []byte{}
		}
	}

	// Convert track name
	trackNameLen := int(cFtn.name.len)
	var trackNameData []byte
	if trackNameLen > 0 {
		trackNameData = C.GoBytes(unsafe.Pointer(&cFtn.name.data[0]), C.int(trackNameLen))
	}

	return FullTrackName{
		Namespace: Namespace{entries: entries},
		TrackName: TrackName{data: trackNameData},
	}
}
