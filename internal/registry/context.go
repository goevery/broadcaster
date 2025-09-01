package registry

import (
	"context"
)

// contextKey is a private type for context keys to avoid collisions
type contextKey string

const (
	connectionInfoKey contextKey = "connection_info"
)

// WithConnectionInfo adds connection information to the context
func WithConnectionInfo(ctx context.Context, conn ConnectionInfo) context.Context {
	return context.WithValue(ctx, connectionInfoKey, conn)
}

// ConnectionInfoFromContext extracts connection information from the context
func ConnectionInfoFromContext(ctx context.Context) (ConnectionInfo, bool) {
	conn, ok := ctx.Value(connectionInfoKey).(ConnectionInfo)
	return conn, ok
}
