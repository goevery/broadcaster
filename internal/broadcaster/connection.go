package broadcaster

import (
	"context"
	"slices"
	"sync"
)

type Authentication struct {
	UserId                string
	AuthorizedChannelsIds []string
}

type Connection struct {
	Id   string
	Send chan Message

	mu             sync.RWMutex
	authentication Authentication
}

func (c *Connection) SetAuthentication(auth Authentication) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.authentication = auth
}

func (c *Connection) GetUserId() string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.authentication.UserId
}

func (c *Connection) IsAuthorized(channelId string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.authentication.UserId == "" {
		return false
	}

	return slices.Contains(c.authentication.AuthorizedChannelsIds, channelId)
}

type contextKey string

const connectionKey contextKey = "connection"

func WithConnection(ctx context.Context, conn *Connection) context.Context {
	return context.WithValue(ctx, connectionKey, conn)
}

func ConnectionFromContext(ctx context.Context) (*Connection, bool) {
	conn, ok := ctx.Value(connectionKey).(*Connection)

	return conn, ok
}
