package persistence

import (
	"context"

	"github.com/juanpmarin/broadcaster/internal/protocol"
)

type Engine interface {
	Setup(ctx context.Context) error
	Save(ctx context.Context, request SaveRequest) (protocol.Message, error)
	List(ctx context.Context, channelId string, lastSeenId string) ([]protocol.Message, error)
}

type SaveRequest struct {
	ChannelId string
	Payload   any
}
