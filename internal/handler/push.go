package handler

import (
	"context"
	"errors"

	"github.com/juanpmarin/broadcaster/internal/persistence"
	"github.com/juanpmarin/broadcaster/internal/protocol"
	"github.com/juanpmarin/broadcaster/internal/registry"
)

type PushRequest struct {
	ChannelId string `json:"channelId"`
	Payload   any    `json:"payload"`
}

type PushHandler struct {
	channelIdValidator   *ChannelIdValidator
	persistenceEngine    persistence.Engine
	subscriptionRegistry registry.Registry
}

func NewPushHandler(
	channelIdValidator *ChannelIdValidator,
	persistenceEngine persistence.Engine,
	subscriptionRegistry registry.Registry,
) *PushHandler {
	return &PushHandler{
		channelIdValidator,
		persistenceEngine,
		subscriptionRegistry,
	}
}

func (h *PushHandler) Handle(ctx context.Context, req PushRequest) (protocol.Message, error) {
	err := h.channelIdValidator.Validate(req.ChannelId)
	if err != nil {
		return protocol.Message{},
			protocol.NewError(protocol.ErrorCodeInvalidArgument, errors.New("invalid channelId"))
	}

	message, err := h.persistenceEngine.Save(ctx, persistence.SaveRequest{
		ChannelId: req.ChannelId,
		Payload:   req.Payload,
	})
	if err != nil {
		return protocol.Message{}, err
	}

	// Broadcast the message to all subscribers of the channel
	err = h.subscriptionRegistry.Broadcast(ctx, req.ChannelId, message)
	if err != nil {
		// Log the error but don't fail the request since message was saved successfully
		// In production, you might want to use structured logging here
	}

	return message, nil
}
