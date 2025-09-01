package registry

import (
	"context"

	"go.uber.org/zap"
)

type Connection interface {
	Id() string
	ClientIp() string
	Send(ctx context.Context, method string, params any) error
}

type Subscription struct {
	Connection Connection
	ChannelId  string
}

// Registry manages active subscriptions for real-time notifications
type Registry interface {
	// Subscribe adds a new subscription for a connection to a channel
	Subscribe(ctx context.Context, channelId string, connection Connection) error

	// Unsubscribe removes a subscription for a connection from a channel
	Unsubscribe(ctx context.Context, channelId string, connectionId string) error

	// Disconnect removes all subscriptions for a connection
	Disconnect(ctx context.Context, connectionId string) error

	// Broadcast sends a message to all subscribers of a channel
	Broadcast(ctx context.Context, channelId string, method string, params any) error
}

type InMemoryRegistry struct {
	logger *zap.Logger
}

func NewInMemoryRegistry(
	logger *zap.Logger,
) *InMemoryRegistry {
	return &InMemoryRegistry{
		logger,
	}
}
