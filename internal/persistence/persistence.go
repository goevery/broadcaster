package persistence

import (
	"context"

	"github.com/juanpmarin/broadcaster/internal/broadcaster"
)

type Engine interface {
	Setup(ctx context.Context) error
	Save(ctx context.Context, request SaveRequest) (broadcaster.Message, error)
	List(ctx context.Context, channelId string, lastSeenId string) ([]broadcaster.Message, error)
}

type SaveRequest struct {
	ChannelId string
	Payload   any
}
