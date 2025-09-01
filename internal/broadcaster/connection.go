package broadcaster

import "context"

type Connection struct {
	Id   string
	Send chan Message
}

type contextKey string

const connectionKey contextKey = "connection"

func WithConnection(ctx context.Context, conn Connection) context.Context {
	return context.WithValue(ctx, connectionKey, conn)
}

func ConnectionFromContext(ctx context.Context) (Connection, bool) {
	conn, ok := ctx.Value(connectionKey).(Connection)

	return conn, ok
}
