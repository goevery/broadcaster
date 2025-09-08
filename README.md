# Broadcaster Protocol Documentation

## Overview

The Broadcaster is a real-time messaging service designed for simplicity and performance. It allows you to build applications that require real-time communication between clients and servers, such as chat applications, live notifications, and collaborative tools.

The protocol is built on top of WebSockets and provides a simple JSON-RPC-like interface for clients to subscribe to channels and receive messages. It also offers a REST API for servers to publish messages to channels.

**Key Features**:

- **Real-time messaging**: Publish messages to clients in real-time.
- **Channel-based communication**: Organize your communication into channels.
- **Secure**: Authenticate clients with JWT and servers with API Keys.
- **Scalable**: Designed to handle a large number of concurrent connections.
- **Simple**: Easy to use and integrate into your existing applications.

## Getting Started

To start using the Broadcaster, you need to:

1.  **Establish a WebSocket connection** to the server.
2.  **Authenticate** the connection using a JWT.
3.  **Subscribe** to one or more channels.
4.  **Start receiving messages**.

To publish messages, you can either use the WebSocket connection (if you have the `publish` scope) or the REST API.

## Authentication

The Broadcaster uses two methods of authentication:

- **JWT (JSON Web Tokens)** for client connections (WebSocket).
- **API Keys** for server-to-server communication (REST API).

### JWT Authentication

Clients must authenticate their WebSocket connection by sending an `auth` request with a valid JWT. The JWT must be signed with the HS256 algorithm and contain the following claims:

- **Standard Claims**:
  - `sub`: User ID (subject)
  - `aud`: Must be `"broadcaster"`
  - `exp`: Expiration time
  - `iat`: Issued at time
- **Custom Claims**:
  - `authorizedChannels`: An array of channel IDs the user is authorized to access.
  - `scope`: An array of strings representing the permissions of the user. Possible values are `"subscribe"` and `"publish"`.

### API Key Authentication

Servers can publish messages to channels using the REST API. To do so, they must include an API Key in the `Authorization` header of their HTTP requests.

**Example**:

```
Authorization: Bearer your-api-key
```

## WebSocket Protocol

### Message Format

The protocol uses a JSON-RPC-inspired message format.

**Request**:

```json
{
  "id": 123,
  "method": "method-name",
  "params": { "foo": "bar" }
}
```

**Response**:

```json
{
  "requestId": 123,
  "result": { "foo": "bar" },
  "error": { "code": "ErrorCode", "message": "Error description" }
}
```

### Connection Lifecycle

1.  **Establish WebSocket connection**.
2.  **Authenticate** with a JWT.
3.  **Subscribe** to desired channels.
4.  **Send/receive** messages.
5.  **Unsubscribe** from channels when no longer needed.
6.  **Close** the WebSocket connection.

### Methods

#### `auth`

Authenticates the connection.

**Params**: `{"token": "jwt-token-string"}`

**Response**: `{"success": true}`

#### `subscribe`

Subscribes the connection to a channel.

**Params**: `{"channel": "channel-name"}`

**Response**: `{"subscriptionId": "sub-123", "timestamp": "2023-01-01T12:00:00Z"}`

#### `unsubscribe`

Unsubscribes the connection from a channel.

**Params**: `{"channel": "channel-name"}`

**Response**: `{"success": true}`

#### `publish`

Publishes a message to a channel. Requires the `publish` scope.

**Params**: `{"channel": "channel-name", "event": "event-name", "payload": {"key": "value"}}`

**Response**: `{"id": "msg-123", "createTime": "2023-01-01T12:00:00Z"}`

#### `heartbeat`

Keeps the connection alive.

**Params**: (none)

**Response**: `{"timestamp": "2023-01-01T12:00:00Z"}`

### Notifications

#### `broadcast`

Sent by the server to clients when a message is published to a channel they are subscribed to.

**Params**: `{"id": "msg-123", "createTime": "2023-01-01T12:00:00Z", "channel": "channel-name", "event": "event-name", "payload": {"key": "value"}}`

## REST API

### `/publish`

Publishes a message to a channel.

**Method**: `POST`

**Headers**: `Authorization: Bearer your-api-key`

**Body**: `{"channel": "channel-name", "event": "event-name", "payload": {"key": "value"}}`

**Response**: `{"id": "msg-123", "createTime": "2023-01-01T12:00:00Z"}`

## Error Handling

Errors are returned in the `error` field of the response message.

**Error Object**:

```json
{
  "code": "ErrorCode",
  "message": "Error description"
}
```

**Error Codes**:

- `InvalidArgument`: Invalid or missing parameters.
- `NotFound`: Requested resource not found.
- `AlreadyExists`: Resource already exists.
- `FailedPrecondition`: Operation cannot be performed in the current state.
- `PermissionDenied`: The caller does not have permission.
- `Unauthenticated`: Authentication is required or has failed.
- `Internal`: Internal server error.

## Authorization Model

- **WebSocket**: Clients must authenticate with a JWT. The `scope` claim in the JWT determines what actions the client can perform.
  - `subscribe`: Allows the client to subscribe to channels and receive messages.
  - `publish`: Allows the client to publish messages to channels.
- **REST API**: Servers must authenticate with an API Key to publish messages.

## Implementation Notes

- The protocol is stateful. The server maintains the authentication and subscription state of each connection.
- Subscriptions are per-connection, not per-user.
- Timestamps are in RFC3339 format in UTC.
