// SPDX-FileCopyrightText: Copyright (c) 2024 Cisco Systems
// SPDX-License-Identifier: BSD-2-Clause

// quicr_shim.h - C shim for libquicr CGO bindings

#ifndef QUICR_SHIM_H
#define QUICR_SHIM_H

#include <stddef.h>
#include <stdint.h>

#ifdef __cplusplus
extern "C" {
#endif

// =============================================================================
// Result/Error Codes
// =============================================================================

typedef enum {
  QUICR_OK = 0,
  QUICR_ERROR_INVALID_PARAM,
  QUICR_ERROR_NOT_CONNECTED,
  QUICR_ERROR_INTERNAL,
  QUICR_ERROR_NOT_READY,
  QUICR_ERROR_TIMEOUT,
  QUICR_ERROR_NOT_AUTHORIZED,
  QUICR_ERROR_ALREADY_EXISTS,
  QUICR_ERROR_NOT_FOUND,
} quicr_result_t;

// =============================================================================
// Client Status
// =============================================================================

typedef enum {
  QUICR_CLIENT_STATUS_READY = 0,
  QUICR_CLIENT_STATUS_NOT_READY,
  QUICR_CLIENT_STATUS_INTERNAL_ERROR,
  QUICR_CLIENT_STATUS_INVALID_PARAMS,
  QUICR_CLIENT_STATUS_CONNECTING,
  QUICR_CLIENT_STATUS_DISCONNECTING,
  QUICR_CLIENT_STATUS_NOT_CONNECTED,
  QUICR_CLIENT_STATUS_FAILED_TO_CONNECT,
  QUICR_CLIENT_STATUS_PENDING_SERVER_SETUP,
} quicr_client_status_t;

// =============================================================================
// Publish Track Status
// =============================================================================

typedef enum {
  QUICR_PUBLISH_STATUS_OK = 0,
  QUICR_PUBLISH_STATUS_NOT_CONNECTED,
  QUICR_PUBLISH_STATUS_NOT_ANNOUNCED,
  QUICR_PUBLISH_STATUS_PENDING_ANNOUNCE_RESPONSE,
  QUICR_PUBLISH_STATUS_ANNOUNCE_NOT_AUTHORIZED,
  QUICR_PUBLISH_STATUS_NO_SUBSCRIBERS,
  QUICR_PUBLISH_STATUS_SENDING_UNANNOUNCE,
  QUICR_PUBLISH_STATUS_SUBSCRIPTION_UPDATED,
  QUICR_PUBLISH_STATUS_NEW_GROUP_REQUESTED,
  QUICR_PUBLISH_STATUS_PENDING_PUBLISH_OK,
  QUICR_PUBLISH_STATUS_PAUSED,
} quicr_publish_status_t;

// =============================================================================
// Publish Object Status (result of publishing)
// =============================================================================

typedef enum {
  QUICR_PUBLISH_OBJECT_OK = 0,
  QUICR_PUBLISH_OBJECT_INTERNAL_ERROR,
  QUICR_PUBLISH_OBJECT_NOT_AUTHORIZED,
  QUICR_PUBLISH_OBJECT_NOT_ANNOUNCED,
  QUICR_PUBLISH_OBJECT_NO_SUBSCRIBERS,
  QUICR_PUBLISH_OBJECT_PAYLOAD_LENGTH_EXCEEDED,
  QUICR_PUBLISH_OBJECT_PREVIOUS_TRUNCATED,
  QUICR_PUBLISH_OBJECT_NO_PREVIOUS,
  QUICR_PUBLISH_OBJECT_DATA_COMPLETE,
  QUICR_PUBLISH_OBJECT_CONTINUATION_NEEDED,
  QUICR_PUBLISH_OBJECT_DATA_INCOMPLETE,
  QUICR_PUBLISH_OBJECT_DATA_TOO_LARGE,
  QUICR_PUBLISH_OBJECT_MUST_START_NEW_GROUP,
  QUICR_PUBLISH_OBJECT_MUST_START_NEW_TRACK,
  QUICR_PUBLISH_OBJECT_PAUSED,
  QUICR_PUBLISH_OBJECT_PENDING_OK,
} quicr_publish_object_status_t;

// =============================================================================
// Subscribe Track Status
// =============================================================================

typedef enum {
  QUICR_SUBSCRIBE_STATUS_OK = 0,
  QUICR_SUBSCRIBE_STATUS_NOT_CONNECTED,
  QUICR_SUBSCRIBE_STATUS_ERROR,
  QUICR_SUBSCRIBE_STATUS_NOT_AUTHORIZED,
  QUICR_SUBSCRIBE_STATUS_NOT_SUBSCRIBED,
  QUICR_SUBSCRIBE_STATUS_PENDING_RESPONSE,
  QUICR_SUBSCRIBE_STATUS_SENDING_UNSUBSCRIBE,
  QUICR_SUBSCRIBE_STATUS_PAUSED,
  QUICR_SUBSCRIBE_STATUS_NEW_GROUP_REQUESTED,
  QUICR_SUBSCRIBE_STATUS_CANCELLED,
  QUICR_SUBSCRIBE_STATUS_DONE_BY_FIN,
  QUICR_SUBSCRIBE_STATUS_DONE_BY_RESET,
} quicr_subscribe_status_t;

// =============================================================================
// Object Status
// =============================================================================

typedef enum {
  QUICR_OBJECT_STATUS_AVAILABLE = 0,
  QUICR_OBJECT_STATUS_DOES_NOT_EXIST = 1,
  QUICR_OBJECT_STATUS_END_OF_GROUP = 3,
  QUICR_OBJECT_STATUS_END_OF_TRACK = 4,
  QUICR_OBJECT_STATUS_END_OF_SUBGROUP = 5,
} quicr_object_status_t;

// =============================================================================
// Track Mode
// =============================================================================

typedef enum {
  QUICR_TRACK_MODE_DATAGRAM = 0,
  QUICR_TRACK_MODE_STREAM = 1,
} quicr_track_mode_t;

// =============================================================================
// Group Order
// =============================================================================

typedef enum {
  QUICR_GROUP_ORDER_ORIGINAL = 0,
  QUICR_GROUP_ORDER_ASCENDING = 1,
  QUICR_GROUP_ORDER_DESCENDING = 2,
} quicr_group_order_t;

// =============================================================================
// Filter Type
// =============================================================================

typedef enum {
  QUICR_FILTER_TYPE_LARGEST_OBJECT = 0,
  QUICR_FILTER_TYPE_LATEST_GROUP = 1,
  QUICR_FILTER_TYPE_LATEST_OBJECT = 2,
  QUICR_FILTER_TYPE_ABSOLUTE_START = 3,
  QUICR_FILTER_TYPE_ABSOLUTE_RANGE = 4,
} quicr_filter_type_t;

// =============================================================================
// Subscribe Namespace Status
// =============================================================================

typedef enum {
  QUICR_SUBSCRIBE_NAMESPACE_STATUS_OK = 0,
  QUICR_SUBSCRIBE_NAMESPACE_STATUS_NOT_SUBSCRIBED,
  QUICR_SUBSCRIBE_NAMESPACE_STATUS_ERROR,
} quicr_subscribe_namespace_status_t;

// =============================================================================
// Publish Namespace Status
// =============================================================================

typedef enum {
  QUICR_PUBLISH_NAMESPACE_STATUS_OK = 0,
  QUICR_PUBLISH_NAMESPACE_STATUS_NOT_CONNECTED,
  QUICR_PUBLISH_NAMESPACE_STATUS_NOT_PUBLISHED,
  QUICR_PUBLISH_NAMESPACE_STATUS_PENDING_RESPONSE,
  QUICR_PUBLISH_NAMESPACE_STATUS_NOT_AUTHORIZED,
  QUICR_PUBLISH_NAMESPACE_STATUS_SENDING_DONE,
  QUICR_PUBLISH_NAMESPACE_STATUS_ERROR,
} quicr_publish_namespace_status_t;

// =============================================================================
// Opaque Handle Types
// =============================================================================

typedef void *quicr_client_t;
typedef void *quicr_publish_track_handler_t;
typedef void *quicr_subscribe_track_handler_t;
typedef void *quicr_publish_namespace_handler_t;
typedef void *quicr_subscribe_namespace_handler_t;

// =============================================================================
// Data Structures
// =============================================================================

#define QUICR_MAX_NAMESPACE_ENTRIES 32
#define QUICR_MAX_ENTRY_SIZE 256
#define QUICR_MAX_TRACK_NAME_SIZE 256
#define QUICR_MAX_ENDPOINT_ID_SIZE 128
#define QUICR_MAX_URI_SIZE 512

// Namespace entry
typedef struct {
  uint8_t data[QUICR_MAX_ENTRY_SIZE];
  uint16_t len;
} quicr_namespace_entry_t;

// Track namespace
typedef struct {
  quicr_namespace_entry_t entries[QUICR_MAX_NAMESPACE_ENTRIES];
  uint8_t num_entries;
} quicr_namespace_t;

// Track name
typedef struct {
  uint8_t data[QUICR_MAX_TRACK_NAME_SIZE];
  uint16_t len;
} quicr_track_name_t;

// Full track name
typedef struct {
  quicr_namespace_t name_space;
  quicr_track_name_t name;
} quicr_full_track_name_t;

// Object headers
typedef struct {
  uint64_t group_id;
  uint64_t object_id;
  uint64_t subgroup_id;
  uint64_t payload_length;
  quicr_object_status_t status;
  uint8_t priority;
  uint16_t ttl;
  quicr_track_mode_t track_mode;
  int has_priority;   // bool: 1 if priority is set
  int has_ttl;        // bool: 1 if ttl is set
  int has_track_mode; // bool: 1 if track_mode is set
} quicr_object_headers_t;

// Received object (headers + data)
typedef struct {
  quicr_object_headers_t headers;
  const uint8_t *data;
  size_t data_len;
} quicr_object_t;

// Client configuration
typedef struct {
  char endpoint_id[QUICR_MAX_ENDPOINT_ID_SIZE];
  char connect_uri[QUICR_MAX_URI_SIZE];
  uint64_t metrics_sample_ms;
  uint64_t tick_service_sleep_delay_us;
} quicr_client_config_t;

// Publish track configuration
typedef struct {
  quicr_full_track_name_t full_track_name;
  quicr_track_mode_t track_mode;
  uint8_t default_priority;
  uint32_t default_ttl;
  int use_announce; // bool
} quicr_publish_track_config_t;

// Subscribe track configuration
typedef struct {
  quicr_full_track_name_t full_track_name;
  uint8_t priority;
  quicr_group_order_t group_order;
  quicr_filter_type_t filter_type;
} quicr_subscribe_track_config_t;

// =============================================================================
// Callback Function Types
// =============================================================================

// Client status changed callback
typedef void (*quicr_client_status_callback_t)(quicr_client_status_t status,
                                               void *user_data);

// Publish track status changed callback
typedef void (*quicr_publish_status_callback_t)(quicr_publish_status_t status,
                                                void *user_data);

// Subscribe track status changed callback
typedef void (*quicr_subscribe_status_callback_t)(
    quicr_subscribe_status_t status, void *user_data);

// Object received callback
typedef void (*quicr_object_received_callback_t)(const quicr_object_t *object,
                                                 void *user_data);

// Publish namespace status changed callback
typedef void (*quicr_publish_namespace_status_callback_t)(
    quicr_publish_namespace_status_t status, void *user_data);

// Subscribe namespace status changed callback
typedef void (*quicr_subscribe_namespace_status_callback_t)(
    quicr_subscribe_namespace_status_t status, void *user_data);

// Subscribe namespace - new track announced callback
typedef void (*quicr_namespace_track_announced_callback_t)(
    const quicr_full_track_name_t *full_track_name, void *user_data);

// =============================================================================
// Client Functions
// =============================================================================

// Initialize default client config
void quicr_client_config_init(quicr_client_config_t *config);

// Create a new client
quicr_client_t quicr_client_create(const quicr_client_config_t *config);

// Destroy client and release resources
void quicr_client_destroy(quicr_client_t client);

// Connect to relay server
quicr_result_t quicr_client_connect(quicr_client_t client);

// Disconnect from relay server
quicr_result_t quicr_client_disconnect(quicr_client_t client);

// Get current client status
quicr_client_status_t quicr_client_get_status(quicr_client_t client);

// Set client status callback
void quicr_client_set_status_callback(quicr_client_t client,
                                      quicr_client_status_callback_t callback,
                                      void *user_data);

// Publish a namespace using handler
quicr_result_t
quicr_client_publish_namespace(quicr_client_t client,
                               quicr_publish_namespace_handler_t handler);

// Publish namespace done using handler
quicr_result_t
quicr_client_publish_namespace_done(quicr_client_t client,
                                    quicr_publish_namespace_handler_t handler);

// Subscribe to a namespace prefix
quicr_result_t
quicr_client_subscribe_namespace(quicr_client_t client,
                                 quicr_subscribe_namespace_handler_t handler);

// Unsubscribe from namespace prefix
quicr_result_t
quicr_client_unsubscribe_namespace(quicr_client_t client,
                                   quicr_subscribe_namespace_handler_t handler);

// =============================================================================
// Publish Track Handler Functions
// =============================================================================

// Create publish track handler
quicr_publish_track_handler_t
quicr_publish_track_handler_create(const quicr_publish_track_config_t *config);

// Destroy publish track handler
void quicr_publish_track_handler_destroy(quicr_publish_track_handler_t handler);

// Register publish track with client
quicr_result_t
quicr_client_publish_track(quicr_client_t client,
                           quicr_publish_track_handler_t handler);

// Unregister publish track
quicr_result_t
quicr_client_unpublish_track(quicr_client_t client,
                             quicr_publish_track_handler_t handler);

// Check if can publish
int quicr_publish_track_handler_can_publish(
    quicr_publish_track_handler_t handler);

// Get publish status
quicr_publish_status_t
quicr_publish_track_handler_get_status(quicr_publish_track_handler_t handler);

// Publish an object
quicr_publish_object_status_t quicr_publish_track_handler_publish_object(
    quicr_publish_track_handler_t handler,
    const quicr_object_headers_t *headers, const uint8_t *data,
    size_t data_len);

// Set default priority
void quicr_publish_track_handler_set_priority(
    quicr_publish_track_handler_t handler, uint8_t priority);

// Set default TTL
void quicr_publish_track_handler_set_ttl(quicr_publish_track_handler_t handler,
                                         uint32_t ttl);

// Set status callback
void quicr_publish_track_handler_set_status_callback(
    quicr_publish_track_handler_t handler,
    quicr_publish_status_callback_t callback, void *user_data);

// =============================================================================
// Subscribe Track Handler Functions
// =============================================================================

// Create subscribe track handler
quicr_subscribe_track_handler_t quicr_subscribe_track_handler_create(
    const quicr_subscribe_track_config_t *config);

// Destroy subscribe track handler
void quicr_subscribe_track_handler_destroy(
    quicr_subscribe_track_handler_t handler);

// Subscribe to a track
quicr_result_t
quicr_client_subscribe_track(quicr_client_t client,
                             quicr_subscribe_track_handler_t handler);

// Unsubscribe from a track
quicr_result_t
quicr_client_unsubscribe_track(quicr_client_t client,
                               quicr_subscribe_track_handler_t handler);

// Get subscribe status
quicr_subscribe_status_t quicr_subscribe_track_handler_get_status(
    quicr_subscribe_track_handler_t handler);

// Set priority
void quicr_subscribe_track_handler_set_priority(
    quicr_subscribe_track_handler_t handler, uint8_t priority);

// Pause subscription
quicr_result_t
quicr_subscribe_track_handler_pause(quicr_subscribe_track_handler_t handler);

// Resume subscription
quicr_result_t
quicr_subscribe_track_handler_resume(quicr_subscribe_track_handler_t handler);

// Set object received callback
void quicr_subscribe_track_handler_set_object_callback(
    quicr_subscribe_track_handler_t handler,
    quicr_object_received_callback_t callback, void *user_data);

// Set status callback
void quicr_subscribe_track_handler_set_status_callback(
    quicr_subscribe_track_handler_t handler,
    quicr_subscribe_status_callback_t callback, void *user_data);

// =============================================================================
// Publish Namespace Handler Functions
// =============================================================================

// Create publish namespace handler
quicr_publish_namespace_handler_t
quicr_publish_namespace_handler_create(const quicr_namespace_t *prefix);

// Destroy publish namespace handler
void quicr_publish_namespace_handler_destroy(
    quicr_publish_namespace_handler_t handler);

// Get publish namespace status
quicr_publish_namespace_status_t
quicr_publish_namespace_handler_get_status(
    quicr_publish_namespace_handler_t handler);

// Set status callback
void quicr_publish_namespace_handler_set_status_callback(
    quicr_publish_namespace_handler_t handler,
    quicr_publish_namespace_status_callback_t callback, void *user_data);

// =============================================================================
// Subscribe Namespace Handler Functions
// =============================================================================

// Create subscribe namespace handler
quicr_subscribe_namespace_handler_t
quicr_subscribe_namespace_handler_create(const quicr_namespace_t *prefix);

// Destroy subscribe namespace handler
void quicr_subscribe_namespace_handler_destroy(
    quicr_subscribe_namespace_handler_t handler);

// Get subscribe namespace status
quicr_subscribe_namespace_status_t
quicr_subscribe_namespace_handler_get_status(
    quicr_subscribe_namespace_handler_t handler);

// Set status callback
void quicr_subscribe_namespace_handler_set_status_callback(
    quicr_subscribe_namespace_handler_t handler,
    quicr_subscribe_namespace_status_callback_t callback, void *user_data);

// Set track announced callback - called when new track is announced under the namespace
void quicr_subscribe_namespace_handler_set_track_announced_callback(
    quicr_subscribe_namespace_handler_t handler,
    quicr_namespace_track_announced_callback_t callback, void *user_data);

// =============================================================================
// Utility Functions
// =============================================================================

// Get string representation of result code
const char *quicr_result_to_string(quicr_result_t result);

// Get string representation of client status
const char *quicr_client_status_to_string(quicr_client_status_t status);

// Get string representation of publish status
const char *quicr_publish_status_to_string(quicr_publish_status_t status);

// Get string representation of subscribe status
const char *quicr_subscribe_status_to_string(quicr_subscribe_status_t status);

#ifdef __cplusplus
}
#endif

#endif // QUICR_SHIM_H
