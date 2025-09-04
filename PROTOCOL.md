# Broadcaster Protocol Documentation

## Overview

The Broadcaster protocol is a real-time messaging system that enables clients to subscribe to channels and receive messages in real-time. It uses a JSON-RPC-like protocol over WebSocket connections with JWT-based authentication and channel-based authorization.

## Connection Flow

1. **WebSocket Connection**: Client establishes a WebSocket connection to the broadcaster server
2. **Authentication**: Client sends an `auth` request with a JWT token
3. **Channel Operations**: Client can join/leave channels and push/receive messages
4. **Heartbeat**: Optional heartbeat mechanism to maintain connection health

## Message Format

The protocol uses a JSON-RPC-inspired message format with two types of messages:

### Request Message

```json
{
  "id": "unique-request-id",     // Optional: present for requests expecting replies
  "method": "method-name",       // Required: the operation to perform
  "params": { ... }              // Optional: method-specific parameters
}
```

### Response Message

```json
{
  "requestId": "original-request-id",  // Required: matches the request ID
  "result": { ... },                   // Optional: success result
  "error": {                           // Optional: error information
    "code": "ErrorCode",
    "message": "Error description",
    "data": { ... }                    // Optional: additional error data
  }
}
```

### Notification Message

Notifications are requests without an `id` field and do not expect a response.

## Error Codes

The protocol defines the following standard error codes:

- `InvalidArgument`: Invalid or missing parameters
- `NotFound`: Requested resource not found
- `AlreadyExists`: Resource already exists
- `FailedPrecondition`: Operation cannot be performed due to current state
- `Unauthenticated`: Authentication required or failed
- `Internal`: Internal server error

## Authentication

### JWT Token Requirements

The JWT token must include:

- **Standard Claims**:
  - `sub`: User ID (subject)
  - `aud`: Must be "broadcaster"
  - `exp`: Expiration time
  - `iat`: Issued at time
- **Custom Claims**:
  - `authorizedChannels`: Array of channel IDs the user can access

### Authentication Method

**Method**: `auth`

**Request Parameters**:

```json
{
  "token": "jwt-token-string"
}
```

**Response**:

```json
{
  "success": true
}
```

**Example**:

```json
{
  "id": "auth-1",
  "method": "auth",
  "params": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
}
```

## Channel Operations

### Channel ID Format

Channel IDs must match the regex pattern: `^([\w-]+:?)*\w$`

Valid examples:

- `general`
- `user-123`
- `team:project:updates`
- `notifications`

### Join Channel

**Method**: `join`

**Request Parameters**:

```json
{
  "channelId": "channel-name",
  "lastSeenMessageId": "message-id" // Optional: for message history recovery
}
```

**Response**:

```json
{
  "subscriptionId": "connection-id",
  "timestamp": "2024-01-15T10:30:00Z",
  "history": [                       // Messages since lastSeenMessageId
    {
      "id": "msg-123",
      "createTime": "2024-01-15T10:25:00Z",
      "channelId": "channel-name",
      "payload": { ... }
    }
  ],
  "historyRecovered": true           // Whether lastSeenMessageId was found
}
```

**Example**:

```json
{
  "id": "join-1",
  "method": "join",
  "params": {
    "channelId": "general",
    "lastSeenMessageId": "msg-456"
  }
}
```

### Leave Channel

**Method**: `leave`

**Request Parameters**:

```json
{
  "channelId": "channel-name"
}
```

**Response**:

```json
{
  "success": true
}
```

**Example**:

```json
{
  "id": "leave-1",
  "method": "leave",
  "params": {
    "channelId": "general"
  }
}
```

### Push Message

**Method**: `push`

**Request Parameters**:

```json
{
  "channelId": "channel-name",
  "payload": { ... }              // Any JSON payload
}
```

**Response**:

```json
{
  "id": "msg-789",
  "createTime": "2024-01-15T10:35:00Z",
  "channelId": "channel-name",
  "payload": { ... }
}
```

**Example**:

```json
{
  "id": "push-1",
  "method": "push",
  "params": {
    "channelId": "general",
    "payload": {
      "type": "chat",
      "message": "Hello, world!",
      "userId": "user-123"
    }
  }
}
```

## Message Broadcasting

When a message is pushed to a channel, it's automatically broadcast to all subscribers of that channel as a notification:

**Notification Format**:

```json
{
  "method": "message",
  "params": {
    "id": "msg-789",
    "createTime": "2024-01-15T10:35:00Z",
    "channelId": "channel-name",
    "payload": { ... }
  }
}
```

## Heartbeat

**Method**: `heartbeat`

**Request**: No parameters required

**Response**:

```json
{
  "timestamp": "2024-01-15T10:40:00Z"
}
```

**Example**:

```json
{
  "id": "ping-1",
  "method": "heartbeat"
}
```

## Authorization Model

### Connection Authentication

- Connections start unauthenticated
- Must authenticate with valid JWT before accessing channels
- Once authenticated, cannot re-authenticate the same connection

### Channel Authorization

- Users can only join channels listed in their JWT `authorizedChannels` claim
- Authorization is checked on every channel operation
- Users can only push messages to channels they're authorized to access

### Message History

- When joining a channel with `lastSeenMessageId`, the server returns:
  - All messages after that ID if the ID is found (`historyRecovered: true`)
  - Empty history if the ID is not found (`historyRecovered: false`)
- History helps clients recover missed messages during reconnections

## Connection Lifecycle

1. **Establish WebSocket connection**
2. **Authenticate with JWT token**
3. **Join desired channels**
4. **Send/receive messages**
5. **Optional heartbeat to maintain connection**
6. **Leave channels when done**
7. **Close WebSocket connection**

## Error Handling

### Common Error Scenarios

1. **Unauthenticated Access**:

   ```json
   {
     "requestId": "join-1",
     "error": {
       "code": "Unauthenticated",
       "message": "user not authorized to access this channel"
     }
   }
   ```

2. **Invalid Channel ID**:

   ```json
   {
     "requestId": "join-1",
     "error": {
       "code": "InvalidArgument",
       "message": "invalid channelId"
     }
   }
   ```

3. **Already Authenticated**:
   ```json
   {
     "requestId": "auth-2",
     "error": {
       "code": "FailedPrecondition",
       "message": "connection is already authenticated"
     }
   }
   ```

## Implementation Notes

- The protocol is stateful - connections maintain authentication and subscription state
- Message persistence is handled server-side, enabling message history recovery
- Channel subscriptions are per-connection, not per-user
- The protocol supports both request-response and notification patterns
- All timestamps use RFC3339 format in UTC
