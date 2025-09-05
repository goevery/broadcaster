package handler

import (
	"context"
	"time"

	"github.com/juanpmarin/broadcaster/internal/broadcaster"
	gonanoid "github.com/matoous/go-nanoid/v2"
)

type PushRequest struct {
	ChannelId string `json:"channelId"`
	Payload   any    `json:"payload"`
}

type PushHandler struct {
	channelIdValidator   *ChannelIdValidator
	subscriptionRegistry broadcaster.Registry
}

func NewPushHandler(
	channelIdValidator *ChannelIdValidator,
	subscriptionRegistry broadcaster.Registry,
) *PushHandler {
	return &PushHandler{
		channelIdValidator,
		subscriptionRegistry,
	}
}

func (h *PushHandler) Handle(ctx context.Context, req PushRequest) (broadcaster.Message, error) {
	err := h.channelIdValidator.Validate(req.ChannelId)
	if err != nil {
		return broadcaster.Message{}, err
	}

	message := broadcaster.Message{
		Id:         gonanoid.Must(),
		CreateTime: time.Now(),
		ChannelId:  req.ChannelId,
		Payload:    req.Payload,
	}

	h.subscriptionRegistry.Broadcast(message)

	return message, nil
}
