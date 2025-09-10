package broadcaster

import (
	"context"
	"sync"

	"github.com/goevery/broadcaster/internal/auth"
)

type Connection struct {
	Id   string
	Send chan Message

	mu             sync.RWMutex
	Seq            uint64
	authentication *auth.Authentication
}

func (c *Connection) NextSeq() uint64 {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.Seq++
	return c.Seq
}

func (c *Connection) SetAuthentication(auth *auth.Authentication) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.authentication = auth
}

func (c *Connection) GetAuthentication() *auth.Authentication {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.authentication
}

func (c *Connection) GetUserId() string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.authentication == nil {
		return ""
	}

	return c.authentication.Subject
}

func (c *Connection) IsAuthorized(channelId string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.authentication == nil {
		return false
	}

	return c.authentication.IsAuthorized(channelId)
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
