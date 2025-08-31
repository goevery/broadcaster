package handler

import (
	"context"
	"errors"

	"github.com/juanpmarin/broadcaster/internal/persistence"
	"github.com/juanpmarin/broadcaster/internal/protocol"
)

type PushRequest struct {
	ChannelId string `json:"channelId"`
	Payload   any    `json:"payload"`
}

type PushHandler struct {
	channelIdValidator *ChannelIdValidator
	persistenceEngine  persistence.Engine
}

func NewPushHandler(
	channelIdValidator *ChannelIdValidator,
	persistenceEngine persistence.Engine,
) *PushHandler {
	return &PushHandler{
		channelIdValidator,
		persistenceEngine,
	}
}

func (h *PushHandler) Handle(ctx context.Context, req PushRequest) (protocol.Message, error) {
	err := h.channelIdValidator.Validate(req.ChannelId)
	if err != nil {
		return protocol.Message{}, NewError(ErrorCodeInvalidArgument, errors.New("invalid channelId"))
	}

	message, err := h.persistenceEngine.Save(ctx, persistence.SaveRequest{
		ChannelId: req.ChannelId,
		Payload:   req.Payload,
	})
	if err != nil {
		return protocol.Message{}, err
	}

	return message, nil
}
