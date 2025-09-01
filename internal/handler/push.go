package handler

import (
	"context"

	"github.com/juanpmarin/broadcaster/internal/broadcaster"
	"github.com/juanpmarin/broadcaster/internal/persistence"
)

type PushRequest struct {
	ChannelId string `json:"channelId"`
	Payload   any    `json:"payload"`
}

type PushHandler struct {
	channelIdValidator   *ChannelIdValidator
	persistenceEngine    persistence.Engine
	subscriptionRegistry broadcaster.Registry
}

func NewPushHandler(
	channelIdValidator *ChannelIdValidator,
	persistenceEngine persistence.Engine,
	subscriptionRegistry broadcaster.Registry,
) *PushHandler {
	return &PushHandler{
		channelIdValidator,
		persistenceEngine,
		subscriptionRegistry,
	}
}

func (h *PushHandler) Handle(ctx context.Context, req PushRequest) (broadcaster.Message, error) {
	err := h.channelIdValidator.Validate(req.ChannelId)
	if err != nil {
		return broadcaster.Message{}, err
	}

	message, err := h.persistenceEngine.Save(ctx, persistence.SaveRequest{
		ChannelId: req.ChannelId,
		Payload:   req.Payload,
	})
	if err != nil {
		return broadcaster.Message{}, err
	}

	h.subscriptionRegistry.Broadcast(message)

	return message, nil
}
