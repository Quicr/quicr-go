// SPDX-FileCopyrightText: Copyright (c) 2024 Cisco Systems
// SPDX-License-Identifier: BSD-2-Clause

// quicr_shim.cpp - C shim implementation for libquicr CGO bindings

#include "quicr_shim.h"

#include <quicr/client.h>
#include <quicr/publish_namespace_handler.h>
#include <quicr/publish_track_handler.h>
#include <quicr/subscribe_namespace_handler.h>
#include <quicr/subscribe_track_handler.h>

#include <cstring>
#include <memory>
#include <mutex>
#include <string>
#include <vector>

// =============================================================================
// Internal Helper Types
// =============================================================================

namespace {

// Custom client class to handle callbacks
class ShimClient : public quicr::Client {
public:
  using Client::Client;

  static std::shared_ptr<ShimClient> Create(const quicr::ClientConfig &cfg) {
    return std::shared_ptr<ShimClient>(new ShimClient(cfg));
  }

  void StatusChanged(quicr::Transport::Status status) override {
    if (status_callback_) {
      status_callback_(ConvertStatus(status), status_user_data_);
    }
  }

  void SetStatusCallback(quicr_client_status_callback_t cb, void *user_data) {
    status_callback_ = cb;
    status_user_data_ = user_data;
  }

  static quicr_client_status_t ConvertStatus(quicr::Transport::Status status) {
    switch (status) {
    case quicr::Transport::Status::kReady:
      return QUICR_CLIENT_STATUS_READY;
    case quicr::Transport::Status::kNotReady:
      return QUICR_CLIENT_STATUS_NOT_READY;
    case quicr::Transport::Status::kInternalError:
      return QUICR_CLIENT_STATUS_INTERNAL_ERROR;
    case quicr::Transport::Status::kInvalidParams:
      return QUICR_CLIENT_STATUS_INVALID_PARAMS;
    case quicr::Transport::Status::kConnecting:
      return QUICR_CLIENT_STATUS_CONNECTING;
    case quicr::Transport::Status::kDisconnecting:
      return QUICR_CLIENT_STATUS_DISCONNECTING;
    case quicr::Transport::Status::kNotConnected:
      return QUICR_CLIENT_STATUS_NOT_CONNECTED;
    case quicr::Transport::Status::kFailedToConnect:
      return QUICR_CLIENT_STATUS_FAILED_TO_CONNECT;
    case quicr::Transport::Status::kPendingServerSetup:
      return QUICR_CLIENT_STATUS_PENDING_SERVER_SETUP;
    default:
      return QUICR_CLIENT_STATUS_INTERNAL_ERROR;
    }
  }

private:
  ShimClient(const quicr::ClientConfig &cfg) : Client(cfg) {}

  quicr_client_status_callback_t status_callback_ = nullptr;
  void *status_user_data_ = nullptr;
};

// Custom publish track handler class
class ShimPublishTrackHandler : public quicr::PublishTrackHandler {
public:
  static std::shared_ptr<ShimPublishTrackHandler>
  Create(const quicr::FullTrackName &full_track_name,
         quicr::TrackMode track_mode, uint8_t default_priority,
         uint32_t default_ttl) {
    return std::shared_ptr<ShimPublishTrackHandler>(new ShimPublishTrackHandler(
        full_track_name, track_mode, default_priority, default_ttl));
  }

  void StatusChanged(Status status) override {
    if (status_callback_) {
      status_callback_(ConvertStatus(status), status_user_data_);
    }
  }

  void SetStatusCallback(quicr_publish_status_callback_t cb, void *user_data) {
    status_callback_ = cb;
    status_user_data_ = user_data;
  }

  static quicr_publish_status_t ConvertStatus(Status status) {
    switch (status) {
    case Status::kOk:
      return QUICR_PUBLISH_STATUS_OK;
    case Status::kNotConnected:
      return QUICR_PUBLISH_STATUS_NOT_CONNECTED;
    case Status::kNotAnnounced:
      return QUICR_PUBLISH_STATUS_NOT_ANNOUNCED;
    case Status::kPendingAnnounceResponse:
      return QUICR_PUBLISH_STATUS_PENDING_ANNOUNCE_RESPONSE;
    case Status::kAnnounceNotAuthorized:
      return QUICR_PUBLISH_STATUS_ANNOUNCE_NOT_AUTHORIZED;
    case Status::kNoSubscribers:
      return QUICR_PUBLISH_STATUS_NO_SUBSCRIBERS;
    case Status::kSendingUnannounce:
      return QUICR_PUBLISH_STATUS_SENDING_UNANNOUNCE;
    case Status::kSubscriptionUpdated:
      return QUICR_PUBLISH_STATUS_SUBSCRIPTION_UPDATED;
    case Status::kNewGroupRequested:
      return QUICR_PUBLISH_STATUS_NEW_GROUP_REQUESTED;
    case Status::kPendingPublishOk:
      return QUICR_PUBLISH_STATUS_PENDING_PUBLISH_OK;
    case Status::kPaused:
      return QUICR_PUBLISH_STATUS_PAUSED;
    default:
      return QUICR_PUBLISH_STATUS_NOT_CONNECTED;
    }
  }

  static quicr_publish_object_status_t
  ConvertPublishObjectStatus(PublishObjectStatus status) {
    switch (status) {
    case PublishObjectStatus::kOk:
      return QUICR_PUBLISH_OBJECT_OK;
    case PublishObjectStatus::kInternalError:
      return QUICR_PUBLISH_OBJECT_INTERNAL_ERROR;
    case PublishObjectStatus::kNotAuthorized:
      return QUICR_PUBLISH_OBJECT_NOT_AUTHORIZED;
    case PublishObjectStatus::kNotAnnounced:
      return QUICR_PUBLISH_OBJECT_NOT_ANNOUNCED;
    case PublishObjectStatus::kNoSubscribers:
      return QUICR_PUBLISH_OBJECT_NO_SUBSCRIBERS;
    case PublishObjectStatus::kObjectPayloadLengthExceeded:
      return QUICR_PUBLISH_OBJECT_PAYLOAD_LENGTH_EXCEEDED;
    case PublishObjectStatus::kPreviousObjectTruncated:
      return QUICR_PUBLISH_OBJECT_PREVIOUS_TRUNCATED;
    case PublishObjectStatus::kNoPreviousObject:
      return QUICR_PUBLISH_OBJECT_NO_PREVIOUS;
    case PublishObjectStatus::kObjectDataComplete:
      return QUICR_PUBLISH_OBJECT_DATA_COMPLETE;
    case PublishObjectStatus::kObjectContinuationDataNeeded:
      return QUICR_PUBLISH_OBJECT_CONTINUATION_NEEDED;
    case PublishObjectStatus::kObjectDataIncomplete:
      return QUICR_PUBLISH_OBJECT_DATA_INCOMPLETE;
    case PublishObjectStatus::kObjectDataTooLarge:
      return QUICR_PUBLISH_OBJECT_DATA_TOO_LARGE;
    case PublishObjectStatus::kPreviousObjectNotCompleteMustStartNewGroup:
      return QUICR_PUBLISH_OBJECT_MUST_START_NEW_GROUP;
    case PublishObjectStatus::kPreviousObjectNotCompleteMustStartNewTrack:
      return QUICR_PUBLISH_OBJECT_MUST_START_NEW_TRACK;
    case PublishObjectStatus::kPaused:
      return QUICR_PUBLISH_OBJECT_PAUSED;
    case PublishObjectStatus::kPendingPublishOk:
      return QUICR_PUBLISH_OBJECT_PENDING_OK;
    default:
      return QUICR_PUBLISH_OBJECT_INTERNAL_ERROR;
    }
  }

private:
  ShimPublishTrackHandler(const quicr::FullTrackName &full_track_name,
                          quicr::TrackMode track_mode, uint8_t default_priority,
                          uint32_t default_ttl)
      : PublishTrackHandler(full_track_name, track_mode, default_priority,
                            default_ttl) {}

  quicr_publish_status_callback_t status_callback_ = nullptr;
  void *status_user_data_ = nullptr;
};

// Custom subscribe track handler class
class ShimSubscribeTrackHandler : public quicr::SubscribeTrackHandler {
public:
  static std::shared_ptr<ShimSubscribeTrackHandler>
  Create(const quicr::FullTrackName &full_track_name, uint8_t priority,
         quicr::messages::GroupOrder group_order) {
    return std::shared_ptr<ShimSubscribeTrackHandler>(
        new ShimSubscribeTrackHandler(full_track_name, priority, group_order));
  }

  void ObjectReceived(const quicr::ObjectHeaders &headers,
                      quicr::BytesSpan data) override {
    if (object_callback_) {
      quicr_object_t obj;
      obj.headers.group_id = headers.group_id;
      obj.headers.object_id = headers.object_id;
      obj.headers.subgroup_id = headers.subgroup_id;
      obj.headers.payload_length = headers.payload_length;
      obj.headers.status = static_cast<quicr_object_status_t>(headers.status);
      obj.headers.priority = headers.priority.value_or(128);
      obj.headers.ttl = headers.ttl.value_or(0);
      obj.headers.has_priority = headers.priority.has_value() ? 1 : 0;
      obj.headers.has_ttl = headers.ttl.has_value() ? 1 : 0;
      obj.headers.has_track_mode = headers.track_mode.has_value() ? 1 : 0;
      if (headers.track_mode.has_value()) {
        obj.headers.track_mode =
            static_cast<quicr_track_mode_t>(headers.track_mode.value());
      }
      obj.data = data.data();
      obj.data_len = data.size();

      object_callback_(&obj, object_user_data_);
    }
  }

  void StatusChanged(Status status) override {
    if (status_callback_) {
      status_callback_(ConvertStatus(status), status_user_data_);
    }
  }

  void SetObjectCallback(quicr_object_received_callback_t cb, void *user_data) {
    object_callback_ = cb;
    object_user_data_ = user_data;
  }

  void SetStatusCallback(quicr_subscribe_status_callback_t cb,
                         void *user_data) {
    status_callback_ = cb;
    status_user_data_ = user_data;
  }

  static quicr_subscribe_status_t ConvertStatus(Status status) {
    switch (status) {
    case Status::kOk:
      return QUICR_SUBSCRIBE_STATUS_OK;
    case Status::kNotConnected:
      return QUICR_SUBSCRIBE_STATUS_NOT_CONNECTED;
    case Status::kError:
      return QUICR_SUBSCRIBE_STATUS_ERROR;
    case Status::kNotAuthorized:
      return QUICR_SUBSCRIBE_STATUS_NOT_AUTHORIZED;
    case Status::kNotSubscribed:
      return QUICR_SUBSCRIBE_STATUS_NOT_SUBSCRIBED;
    case Status::kPendingResponse:
      return QUICR_SUBSCRIBE_STATUS_PENDING_RESPONSE;
    case Status::kSendingUnsubscribe:
      return QUICR_SUBSCRIBE_STATUS_SENDING_UNSUBSCRIBE;
    case Status::kPaused:
      return QUICR_SUBSCRIBE_STATUS_PAUSED;
    case Status::kNewGroupRequested:
      return QUICR_SUBSCRIBE_STATUS_NEW_GROUP_REQUESTED;
    case Status::kCancelled:
      return QUICR_SUBSCRIBE_STATUS_CANCELLED;
    case Status::kDoneByFin:
      return QUICR_SUBSCRIBE_STATUS_DONE_BY_FIN;
    case Status::kDoneByReset:
      return QUICR_SUBSCRIBE_STATUS_DONE_BY_RESET;
    default:
      return QUICR_SUBSCRIBE_STATUS_ERROR;
    }
  }

private:
  ShimSubscribeTrackHandler(const quicr::FullTrackName &full_track_name,
                            uint8_t priority,
                            quicr::messages::GroupOrder group_order)
      : SubscribeTrackHandler(full_track_name, priority, group_order) {}

  quicr_object_received_callback_t object_callback_ = nullptr;
  void *object_user_data_ = nullptr;
  quicr_subscribe_status_callback_t status_callback_ = nullptr;
  void *status_user_data_ = nullptr;
};

// Custom publish namespace handler class
class ShimPublishNamespaceHandler : public quicr::PublishNamespaceHandler {
public:
  static std::shared_ptr<ShimPublishNamespaceHandler>
  Create(const quicr::TrackNamespace &prefix) {
    return std::shared_ptr<ShimPublishNamespaceHandler>(
        new ShimPublishNamespaceHandler(prefix));
  }

  void StatusChanged(Status status) override {
    if (status_callback_) {
      status_callback_(ConvertStatus(status), status_user_data_);
    }
  }

  void SetStatusCallback(quicr_publish_namespace_status_callback_t cb,
                         void *user_data) {
    status_callback_ = cb;
    status_user_data_ = user_data;
  }

  static quicr_publish_namespace_status_t ConvertStatus(Status status) {
    switch (status) {
    case Status::kOk:
      return QUICR_PUBLISH_NAMESPACE_STATUS_OK;
    case Status::kNotConnected:
      return QUICR_PUBLISH_NAMESPACE_STATUS_NOT_CONNECTED;
    case Status::kNotPublished:
      return QUICR_PUBLISH_NAMESPACE_STATUS_NOT_PUBLISHED;
    case Status::kPendingResponse:
      return QUICR_PUBLISH_NAMESPACE_STATUS_PENDING_RESPONSE;
    case Status::kPublishNotAuthorized:
      return QUICR_PUBLISH_NAMESPACE_STATUS_NOT_AUTHORIZED;
    case Status::kSendingDone:
      return QUICR_PUBLISH_NAMESPACE_STATUS_SENDING_DONE;
    case Status::kError:
      return QUICR_PUBLISH_NAMESPACE_STATUS_ERROR;
    default:
      return QUICR_PUBLISH_NAMESPACE_STATUS_ERROR;
    }
  }

private:
  ShimPublishNamespaceHandler(const quicr::TrackNamespace &prefix)
      : PublishNamespaceHandler(prefix) {}

  quicr_publish_namespace_status_callback_t status_callback_ = nullptr;
  void *status_user_data_ = nullptr;
};

// Custom subscribe namespace handler class
class ShimSubscribeNamespaceHandler : public quicr::SubscribeNamespaceHandler {
public:
  static std::shared_ptr<ShimSubscribeNamespaceHandler>
  Create(const quicr::TrackNamespace &prefix) {
    return std::shared_ptr<ShimSubscribeNamespaceHandler>(
        new ShimSubscribeNamespaceHandler(prefix));
  }

  void StatusChanged(Status status) override {
    if (status_callback_) {
      status_callback_(ConvertStatus(status), status_user_data_);
    }
  }

  void SetStatusCallback(quicr_subscribe_namespace_status_callback_t cb,
                         void *user_data) {
    status_callback_ = cb;
    status_user_data_ = user_data;
  }

  void SetTrackAnnouncedCallback(quicr_namespace_track_announced_callback_t cb,
                                 void *user_data) {
    track_announced_callback_ = cb;
    track_announced_user_data_ = user_data;
  }

  static quicr_subscribe_namespace_status_t ConvertStatus(Status status) {
    switch (status) {
    case Status::kOk:
      return QUICR_SUBSCRIBE_NAMESPACE_STATUS_OK;
    case Status::kNotSubscribed:
      return QUICR_SUBSCRIBE_NAMESPACE_STATUS_NOT_SUBSCRIBED;
    case Status::kError:
      return QUICR_SUBSCRIBE_NAMESPACE_STATUS_ERROR;
    default:
      return QUICR_SUBSCRIBE_NAMESPACE_STATUS_ERROR;
    }
  }

private:
  ShimSubscribeNamespaceHandler(const quicr::TrackNamespace &prefix)
      : SubscribeNamespaceHandler(prefix) {}

  quicr_subscribe_namespace_status_callback_t status_callback_ = nullptr;
  void *status_user_data_ = nullptr;
  quicr_namespace_track_announced_callback_t track_announced_callback_ =
      nullptr;
  void *track_announced_user_data_ = nullptr;
};

// Context wrapper to hold shared_ptr
struct ClientContext {
  std::shared_ptr<ShimClient> client;
};

struct PublishHandlerContext {
  std::shared_ptr<ShimPublishTrackHandler> handler;
};

struct SubscribeHandlerContext {
  std::shared_ptr<ShimSubscribeTrackHandler> handler;
};

struct PublishNamespaceHandlerContext {
  std::shared_ptr<ShimPublishNamespaceHandler> handler;
};

struct SubscribeNamespaceHandlerContext {
  std::shared_ptr<ShimSubscribeNamespaceHandler> handler;
};

// Conversion helpers
quicr::TrackNamespace ConvertNamespace(const quicr_namespace_t *ns) {
  std::vector<std::vector<uint8_t>> entries;
  for (uint8_t i = 0; i < ns->num_entries; ++i) {
    const auto &entry = ns->entries[i];
    entries.emplace_back(entry.data, entry.data + entry.len);
  }
  return quicr::TrackNamespace(entries);
}

quicr::FullTrackName ConvertFullTrackName(const quicr_full_track_name_t *ftn) {
  quicr::FullTrackName result;
  result.name_space = ConvertNamespace(&ftn->name_space);
  result.name =
      std::vector<uint8_t>(ftn->name.data, ftn->name.data + ftn->name.len);
  return result;
}

quicr::TrackMode ConvertTrackMode(quicr_track_mode_t mode) {
  switch (mode) {
  case QUICR_TRACK_MODE_DATAGRAM:
    return quicr::TrackMode::kDatagram;
  case QUICR_TRACK_MODE_STREAM:
  default:
    return quicr::TrackMode::kStream;
  }
}

quicr::messages::GroupOrder ConvertGroupOrder(quicr_group_order_t order) {
  switch (order) {
  case QUICR_GROUP_ORDER_ASCENDING:
    return quicr::messages::GroupOrder::kAscending;
  case QUICR_GROUP_ORDER_DESCENDING:
    return quicr::messages::GroupOrder::kDescending;
  case QUICR_GROUP_ORDER_ORIGINAL:
  default:
    return quicr::messages::GroupOrder::kOriginalPublisherOrder;
  }
}

// Note: FilterType in libquicr has changed significantly. The new API uses
// a Filter variant type. For simplicity, we use the default filter (no filter)
// and group order is the main control for subscription behavior.

quicr::ObjectHeaders
ConvertObjectHeaders(const quicr_object_headers_t *headers) {
  quicr::ObjectHeaders result;
  result.group_id = headers->group_id;
  result.object_id = headers->object_id;
  result.subgroup_id = headers->subgroup_id;
  result.payload_length = headers->payload_length;
  result.status = static_cast<quicr::ObjectStatus>(headers->status);
  if (headers->has_priority) {
    result.priority = headers->priority;
  }
  if (headers->has_ttl) {
    result.ttl = headers->ttl;
  }
  if (headers->has_track_mode) {
    result.track_mode = ConvertTrackMode(headers->track_mode);
  }
  return result;
}

} // anonymous namespace

// =============================================================================
// C API Implementation
// =============================================================================

extern "C" {

// Client config initialization
void quicr_client_config_init(quicr_client_config_t *config) {
  if (!config)
    return;
  std::memset(config, 0, sizeof(*config));
  std::strncpy(config->endpoint_id, "quicr-go-client",
               QUICR_MAX_ENDPOINT_ID_SIZE - 1);
  config->metrics_sample_ms = 5000;
  config->tick_service_sleep_delay_us = 333;
}

// Create client
quicr_client_t quicr_client_create(const quicr_client_config_t *config) {
  if (!config)
    return nullptr;

  try {
    quicr::ClientConfig cfg;
    cfg.endpoint_id = config->endpoint_id;
    cfg.connect_uri = config->connect_uri;
    cfg.metrics_sample_ms = config->metrics_sample_ms;
    cfg.tick_service_sleep_delay_us = config->tick_service_sleep_delay_us;

    auto ctx = new ClientContext();
    ctx->client = ShimClient::Create(cfg);
    return static_cast<quicr_client_t>(ctx);
  } catch (...) {
    return nullptr;
  }
}

// Destroy client
void quicr_client_destroy(quicr_client_t client) {
  if (!client)
    return;
  auto ctx = static_cast<ClientContext *>(client);
  delete ctx;
}

// Connect
quicr_result_t quicr_client_connect(quicr_client_t client) {
  if (!client)
    return QUICR_ERROR_INVALID_PARAM;
  auto ctx = static_cast<ClientContext *>(client);

  try {
    auto status = ctx->client->Connect();
    if (status == quicr::Transport::Status::kConnecting ||
        status == quicr::Transport::Status::kReady) {
      return QUICR_OK;
    }
    return QUICR_ERROR_INTERNAL;
  } catch (...) {
    return QUICR_ERROR_INTERNAL;
  }
}

// Disconnect
quicr_result_t quicr_client_disconnect(quicr_client_t client) {
  if (!client)
    return QUICR_ERROR_INVALID_PARAM;
  auto ctx = static_cast<ClientContext *>(client);

  try {
    ctx->client->Disconnect();
    return QUICR_OK;
  } catch (...) {
    return QUICR_ERROR_INTERNAL;
  }
}

// Get status
quicr_client_status_t quicr_client_get_status(quicr_client_t client) {
  if (!client)
    return QUICR_CLIENT_STATUS_NOT_CONNECTED;
  auto ctx = static_cast<ClientContext *>(client);
  return ShimClient::ConvertStatus(ctx->client->GetStatus());
}

// Set status callback
void quicr_client_set_status_callback(quicr_client_t client,
                                      quicr_client_status_callback_t callback,
                                      void *user_data) {
  if (!client)
    return;
  auto ctx = static_cast<ClientContext *>(client);
  ctx->client->SetStatusCallback(callback, user_data);
}

// Publish namespace using handler
quicr_result_t
quicr_client_publish_namespace(quicr_client_t client,
                               quicr_publish_namespace_handler_t handler) {
  if (!client || !handler)
    return QUICR_ERROR_INVALID_PARAM;
  auto ctx = static_cast<ClientContext *>(client);
  auto handler_ctx = static_cast<PublishNamespaceHandlerContext *>(handler);

  try {
    ctx->client->PublishNamespace(handler_ctx->handler);
    return QUICR_OK;
  } catch (...) {
    return QUICR_ERROR_INTERNAL;
  }
}

// Publish namespace done using handler
quicr_result_t
quicr_client_publish_namespace_done(quicr_client_t client,
                                    quicr_publish_namespace_handler_t handler) {
  if (!client || !handler)
    return QUICR_ERROR_INVALID_PARAM;
  auto ctx = static_cast<ClientContext *>(client);
  auto handler_ctx = static_cast<PublishNamespaceHandlerContext *>(handler);

  try {
    ctx->client->PublishNamespaceDone(handler_ctx->handler);
    return QUICR_OK;
  } catch (...) {
    return QUICR_ERROR_INTERNAL;
  }
}

// Subscribe to namespace prefix
quicr_result_t
quicr_client_subscribe_namespace(quicr_client_t client,
                                 quicr_subscribe_namespace_handler_t handler) {
  if (!client || !handler)
    return QUICR_ERROR_INVALID_PARAM;
  auto ctx = static_cast<ClientContext *>(client);
  auto handler_ctx = static_cast<SubscribeNamespaceHandlerContext *>(handler);

  try {
    ctx->client->SubscribeNamespace(handler_ctx->handler);
    return QUICR_OK;
  } catch (...) {
    return QUICR_ERROR_INTERNAL;
  }
}

// Unsubscribe from namespace prefix
quicr_result_t quicr_client_unsubscribe_namespace(
    quicr_client_t client, quicr_subscribe_namespace_handler_t handler) {
  if (!client || !handler)
    return QUICR_ERROR_INVALID_PARAM;
  auto ctx = static_cast<ClientContext *>(client);
  auto handler_ctx = static_cast<SubscribeNamespaceHandlerContext *>(handler);

  try {
    ctx->client->UnsubscribeNamespace(handler_ctx->handler);
    return QUICR_OK;
  } catch (...) {
    return QUICR_ERROR_INTERNAL;
  }
}

// =============================================================================
// Publish Track Handler
// =============================================================================

quicr_publish_track_handler_t
quicr_publish_track_handler_create(const quicr_publish_track_config_t *config) {
  if (!config)
    return nullptr;

  try {
    auto ftn = ConvertFullTrackName(&config->full_track_name);
    auto track_mode = ConvertTrackMode(config->track_mode);

    auto ctx = new PublishHandlerContext();
    ctx->handler = ShimPublishTrackHandler::Create(
        ftn, track_mode, config->default_priority, config->default_ttl);
    // Note: use_announce is now handled via PublishNamespace() instead of per-track
    return static_cast<quicr_publish_track_handler_t>(ctx);
  } catch (...) {
    return nullptr;
  }
}

void quicr_publish_track_handler_destroy(
    quicr_publish_track_handler_t handler) {
  if (!handler)
    return;
  auto ctx = static_cast<PublishHandlerContext *>(handler);
  delete ctx;
}

quicr_result_t
quicr_client_publish_track(quicr_client_t client,
                           quicr_publish_track_handler_t handler) {
  if (!client || !handler)
    return QUICR_ERROR_INVALID_PARAM;

  auto client_ctx = static_cast<ClientContext *>(client);
  auto handler_ctx = static_cast<PublishHandlerContext *>(handler);

  try {
    client_ctx->client->PublishTrack(handler_ctx->handler);
    return QUICR_OK;
  } catch (...) {
    return QUICR_ERROR_INTERNAL;
  }
}

quicr_result_t
quicr_client_unpublish_track(quicr_client_t client,
                             quicr_publish_track_handler_t handler) {
  if (!client || !handler)
    return QUICR_ERROR_INVALID_PARAM;

  auto client_ctx = static_cast<ClientContext *>(client);
  auto handler_ctx = static_cast<PublishHandlerContext *>(handler);

  try {
    client_ctx->client->UnpublishTrack(handler_ctx->handler);
    return QUICR_OK;
  } catch (...) {
    return QUICR_ERROR_INTERNAL;
  }
}

int quicr_publish_track_handler_can_publish(
    quicr_publish_track_handler_t handler) {
  if (!handler)
    return 0;
  auto ctx = static_cast<PublishHandlerContext *>(handler);
  return ctx->handler->CanPublish() ? 1 : 0;
}

quicr_publish_status_t
quicr_publish_track_handler_get_status(quicr_publish_track_handler_t handler) {
  if (!handler)
    return QUICR_PUBLISH_STATUS_NOT_CONNECTED;
  auto ctx = static_cast<PublishHandlerContext *>(handler);
  return ShimPublishTrackHandler::ConvertStatus(ctx->handler->GetStatus());
}

quicr_publish_object_status_t quicr_publish_track_handler_publish_object(
    quicr_publish_track_handler_t handler,
    const quicr_object_headers_t *headers, const uint8_t *data,
    size_t data_len) {
  if (!handler || !headers)
    return QUICR_PUBLISH_OBJECT_INTERNAL_ERROR;

  auto ctx = static_cast<PublishHandlerContext *>(handler);

  try {
    auto obj_headers = ConvertObjectHeaders(headers);
    obj_headers.payload_length = data_len;
    quicr::BytesSpan span(data, data_len);

    auto status = ctx->handler->PublishObject(obj_headers, span);
    return ShimPublishTrackHandler::ConvertPublishObjectStatus(status);
  } catch (...) {
    return QUICR_PUBLISH_OBJECT_INTERNAL_ERROR;
  }
}

void quicr_publish_track_handler_set_priority(
    quicr_publish_track_handler_t handler, uint8_t priority) {
  if (!handler)
    return;
  auto ctx = static_cast<PublishHandlerContext *>(handler);
  ctx->handler->SetDefaultPriority(priority);
}

void quicr_publish_track_handler_set_ttl(quicr_publish_track_handler_t handler,
                                         uint32_t ttl) {
  if (!handler)
    return;
  auto ctx = static_cast<PublishHandlerContext *>(handler);
  ctx->handler->SetDefaultTTL(ttl);
}

void quicr_publish_track_handler_set_status_callback(
    quicr_publish_track_handler_t handler,
    quicr_publish_status_callback_t callback, void *user_data) {
  if (!handler)
    return;
  auto ctx = static_cast<PublishHandlerContext *>(handler);
  ctx->handler->SetStatusCallback(callback, user_data);
}

// =============================================================================
// Subscribe Track Handler
// =============================================================================

quicr_subscribe_track_handler_t quicr_subscribe_track_handler_create(
    const quicr_subscribe_track_config_t *config) {
  if (!config)
    return nullptr;

  try {
    auto ftn = ConvertFullTrackName(&config->full_track_name);
    auto group_order = ConvertGroupOrder(config->group_order);

    auto ctx = new SubscribeHandlerContext();
    ctx->handler =
        ShimSubscribeTrackHandler::Create(ftn, config->priority, group_order);
    return static_cast<quicr_subscribe_track_handler_t>(ctx);
  } catch (...) {
    return nullptr;
  }
}

void quicr_subscribe_track_handler_destroy(
    quicr_subscribe_track_handler_t handler) {
  if (!handler)
    return;
  auto ctx = static_cast<SubscribeHandlerContext *>(handler);
  delete ctx;
}

quicr_result_t
quicr_client_subscribe_track(quicr_client_t client,
                             quicr_subscribe_track_handler_t handler) {
  if (!client || !handler)
    return QUICR_ERROR_INVALID_PARAM;

  auto client_ctx = static_cast<ClientContext *>(client);
  auto handler_ctx = static_cast<SubscribeHandlerContext *>(handler);

  try {
    client_ctx->client->SubscribeTrack(handler_ctx->handler);
    return QUICR_OK;
  } catch (...) {
    return QUICR_ERROR_INTERNAL;
  }
}

quicr_result_t
quicr_client_unsubscribe_track(quicr_client_t client,
                               quicr_subscribe_track_handler_t handler) {
  if (!client || !handler)
    return QUICR_ERROR_INVALID_PARAM;

  auto client_ctx = static_cast<ClientContext *>(client);
  auto handler_ctx = static_cast<SubscribeHandlerContext *>(handler);

  try {
    client_ctx->client->UnsubscribeTrack(handler_ctx->handler);
    return QUICR_OK;
  } catch (...) {
    return QUICR_ERROR_INTERNAL;
  }
}

quicr_subscribe_status_t quicr_subscribe_track_handler_get_status(
    quicr_subscribe_track_handler_t handler) {
  if (!handler)
    return QUICR_SUBSCRIBE_STATUS_NOT_CONNECTED;
  auto ctx = static_cast<SubscribeHandlerContext *>(handler);
  return ShimSubscribeTrackHandler::ConvertStatus(ctx->handler->GetStatus());
}

void quicr_subscribe_track_handler_set_priority(
    quicr_subscribe_track_handler_t handler, uint8_t priority) {
  if (!handler)
    return;
  auto ctx = static_cast<SubscribeHandlerContext *>(handler);
  ctx->handler->SetPriority(priority);
}

quicr_result_t
quicr_subscribe_track_handler_pause(quicr_subscribe_track_handler_t handler) {
  if (!handler)
    return QUICR_ERROR_INVALID_PARAM;
  auto ctx = static_cast<SubscribeHandlerContext *>(handler);

  try {
    ctx->handler->Pause();
    return QUICR_OK;
  } catch (...) {
    return QUICR_ERROR_INTERNAL;
  }
}

quicr_result_t
quicr_subscribe_track_handler_resume(quicr_subscribe_track_handler_t handler) {
  if (!handler)
    return QUICR_ERROR_INVALID_PARAM;
  auto ctx = static_cast<SubscribeHandlerContext *>(handler);

  try {
    ctx->handler->Resume();
    return QUICR_OK;
  } catch (...) {
    return QUICR_ERROR_INTERNAL;
  }
}

void quicr_subscribe_track_handler_set_object_callback(
    quicr_subscribe_track_handler_t handler,
    quicr_object_received_callback_t callback, void *user_data) {
  if (!handler)
    return;
  auto ctx = static_cast<SubscribeHandlerContext *>(handler);
  ctx->handler->SetObjectCallback(callback, user_data);
}

void quicr_subscribe_track_handler_set_status_callback(
    quicr_subscribe_track_handler_t handler,
    quicr_subscribe_status_callback_t callback, void *user_data) {
  if (!handler)
    return;
  auto ctx = static_cast<SubscribeHandlerContext *>(handler);
  ctx->handler->SetStatusCallback(callback, user_data);
}

// =============================================================================
// Publish Namespace Handler
// =============================================================================

quicr_publish_namespace_handler_t
quicr_publish_namespace_handler_create(const quicr_namespace_t *prefix) {
  if (!prefix)
    return nullptr;

  try {
    auto ns = ConvertNamespace(prefix);
    auto ctx = new PublishNamespaceHandlerContext();
    ctx->handler = ShimPublishNamespaceHandler::Create(ns);
    return static_cast<quicr_publish_namespace_handler_t>(ctx);
  } catch (...) {
    return nullptr;
  }
}

void quicr_publish_namespace_handler_destroy(
    quicr_publish_namespace_handler_t handler) {
  if (!handler)
    return;
  auto ctx = static_cast<PublishNamespaceHandlerContext *>(handler);
  delete ctx;
}

quicr_publish_namespace_status_t quicr_publish_namespace_handler_get_status(
    quicr_publish_namespace_handler_t handler) {
  if (!handler)
    return QUICR_PUBLISH_NAMESPACE_STATUS_NOT_CONNECTED;
  auto ctx = static_cast<PublishNamespaceHandlerContext *>(handler);
  return ShimPublishNamespaceHandler::ConvertStatus(ctx->handler->GetStatus());
}

void quicr_publish_namespace_handler_set_status_callback(
    quicr_publish_namespace_handler_t handler,
    quicr_publish_namespace_status_callback_t callback, void *user_data) {
  if (!handler)
    return;
  auto ctx = static_cast<PublishNamespaceHandlerContext *>(handler);
  ctx->handler->SetStatusCallback(callback, user_data);
}

// =============================================================================
// Subscribe Namespace Handler
// =============================================================================

quicr_subscribe_namespace_handler_t
quicr_subscribe_namespace_handler_create(const quicr_namespace_t *prefix) {
  if (!prefix)
    return nullptr;

  try {
    auto ns = ConvertNamespace(prefix);
    auto ctx = new SubscribeNamespaceHandlerContext();
    ctx->handler = ShimSubscribeNamespaceHandler::Create(ns);
    return static_cast<quicr_subscribe_namespace_handler_t>(ctx);
  } catch (...) {
    return nullptr;
  }
}

void quicr_subscribe_namespace_handler_destroy(
    quicr_subscribe_namespace_handler_t handler) {
  if (!handler)
    return;
  auto ctx = static_cast<SubscribeNamespaceHandlerContext *>(handler);
  delete ctx;
}

quicr_subscribe_namespace_status_t quicr_subscribe_namespace_handler_get_status(
    quicr_subscribe_namespace_handler_t handler) {
  if (!handler)
    return QUICR_SUBSCRIBE_NAMESPACE_STATUS_NOT_SUBSCRIBED;
  auto ctx = static_cast<SubscribeNamespaceHandlerContext *>(handler);
  return ShimSubscribeNamespaceHandler::ConvertStatus(
      ctx->handler->GetStatus());
}

void quicr_subscribe_namespace_handler_set_status_callback(
    quicr_subscribe_namespace_handler_t handler,
    quicr_subscribe_namespace_status_callback_t callback, void *user_data) {
  if (!handler)
    return;
  auto ctx = static_cast<SubscribeNamespaceHandlerContext *>(handler);
  ctx->handler->SetStatusCallback(callback, user_data);
}

void quicr_subscribe_namespace_handler_set_track_announced_callback(
    quicr_subscribe_namespace_handler_t handler,
    quicr_namespace_track_announced_callback_t callback, void *user_data) {
  if (!handler)
    return;
  auto ctx = static_cast<SubscribeNamespaceHandlerContext *>(handler);
  ctx->handler->SetTrackAnnouncedCallback(callback, user_data);
}

// =============================================================================
// Utility Functions
// =============================================================================

const char *quicr_result_to_string(quicr_result_t result) {
  switch (result) {
  case QUICR_OK:
    return "OK";
  case QUICR_ERROR_INVALID_PARAM:
    return "Invalid parameter";
  case QUICR_ERROR_NOT_CONNECTED:
    return "Not connected";
  case QUICR_ERROR_INTERNAL:
    return "Internal error";
  case QUICR_ERROR_NOT_READY:
    return "Not ready";
  case QUICR_ERROR_TIMEOUT:
    return "Timeout";
  case QUICR_ERROR_NOT_AUTHORIZED:
    return "Not authorized";
  case QUICR_ERROR_ALREADY_EXISTS:
    return "Already exists";
  case QUICR_ERROR_NOT_FOUND:
    return "Not found";
  default:
    return "Unknown error";
  }
}

const char *quicr_client_status_to_string(quicr_client_status_t status) {
  switch (status) {
  case QUICR_CLIENT_STATUS_READY:
    return "Ready";
  case QUICR_CLIENT_STATUS_NOT_READY:
    return "Not ready";
  case QUICR_CLIENT_STATUS_INTERNAL_ERROR:
    return "Internal error";
  case QUICR_CLIENT_STATUS_INVALID_PARAMS:
    return "Invalid parameters";
  case QUICR_CLIENT_STATUS_CONNECTING:
    return "Connecting";
  case QUICR_CLIENT_STATUS_DISCONNECTING:
    return "Disconnecting";
  case QUICR_CLIENT_STATUS_NOT_CONNECTED:
    return "Not connected";
  case QUICR_CLIENT_STATUS_FAILED_TO_CONNECT:
    return "Failed to connect";
  case QUICR_CLIENT_STATUS_PENDING_SERVER_SETUP:
    return "Pending server setup";
  default:
    return "Unknown status";
  }
}

const char *quicr_publish_status_to_string(quicr_publish_status_t status) {
  switch (status) {
  case QUICR_PUBLISH_STATUS_OK:
    return "OK";
  case QUICR_PUBLISH_STATUS_NOT_CONNECTED:
    return "Not connected";
  case QUICR_PUBLISH_STATUS_NOT_ANNOUNCED:
    return "Not announced";
  case QUICR_PUBLISH_STATUS_PENDING_ANNOUNCE_RESPONSE:
    return "Pending announce response";
  case QUICR_PUBLISH_STATUS_ANNOUNCE_NOT_AUTHORIZED:
    return "Announce not authorized";
  case QUICR_PUBLISH_STATUS_NO_SUBSCRIBERS:
    return "No subscribers";
  case QUICR_PUBLISH_STATUS_SENDING_UNANNOUNCE:
    return "Sending unannounce";
  case QUICR_PUBLISH_STATUS_SUBSCRIPTION_UPDATED:
    return "Subscription updated";
  case QUICR_PUBLISH_STATUS_NEW_GROUP_REQUESTED:
    return "New group requested";
  case QUICR_PUBLISH_STATUS_PENDING_PUBLISH_OK:
    return "Pending publish OK";
  case QUICR_PUBLISH_STATUS_PAUSED:
    return "Paused";
  default:
    return "Unknown status";
  }
}

const char *quicr_subscribe_status_to_string(quicr_subscribe_status_t status) {
  switch (status) {
  case QUICR_SUBSCRIBE_STATUS_OK:
    return "OK";
  case QUICR_SUBSCRIBE_STATUS_NOT_CONNECTED:
    return "Not connected";
  case QUICR_SUBSCRIBE_STATUS_ERROR:
    return "Error";
  case QUICR_SUBSCRIBE_STATUS_NOT_AUTHORIZED:
    return "Not authorized";
  case QUICR_SUBSCRIBE_STATUS_NOT_SUBSCRIBED:
    return "Not subscribed";
  case QUICR_SUBSCRIBE_STATUS_PENDING_RESPONSE:
    return "Pending response";
  case QUICR_SUBSCRIBE_STATUS_SENDING_UNSUBSCRIBE:
    return "Sending unsubscribe";
  case QUICR_SUBSCRIBE_STATUS_PAUSED:
    return "Paused";
  case QUICR_SUBSCRIBE_STATUS_NEW_GROUP_REQUESTED:
    return "New group requested";
  case QUICR_SUBSCRIBE_STATUS_CANCELLED:
    return "Cancelled";
  case QUICR_SUBSCRIBE_STATUS_DONE_BY_FIN:
    return "Done by FIN";
  case QUICR_SUBSCRIBE_STATUS_DONE_BY_RESET:
    return "Done by reset";
  default:
    return "Unknown status";
  }
}

} // extern "C"
