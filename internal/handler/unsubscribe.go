package handler

import (
	"context"
	"errors"

	"github.com/goevery/broadcaster/internal/broadcaster"
)

type UnsubscribeRequest struct {
	Channel string `json:"channel"`
}

type UnsubscribeResponse struct {
	Success bool `json:"success"`
}

type UnsubscribeHandlerInterface interface {
	Handle(ctx context.Context, req UnsubscribeRequest) (UnsubscribeResponse, error)
}

type UnsubscribeHandler struct {
	channelValidator     *ChannelValidator
	subscriptionRegistry broadcaster.Registry
}

func NewUnsubscribeHandler(
	channelValidator *ChannelValidator,
	subscriptionRegistry broadcaster.Registry,
) *UnsubscribeHandler {
	return &UnsubscribeHandler{
		channelValidator,
		subscriptionRegistry,
	}
}

func (h *UnsubscribeHandler) Handle(ctx context.Context, req UnsubscribeRequest) (UnsubscribeResponse, error) {
	err := h.channelValidator.Validate(req.Channel)
	if err != nil {
		return UnsubscribeResponse{}, err
	}

	connection, ok := broadcaster.ConnectionFromContext(ctx)
	if !ok {
		return UnsubscribeResponse{}, errors.New("connection not found in context")
	}

	h.subscriptionRegistry.Unsubscribe(req.Channel, connection.Id)

	return UnsubscribeResponse{
		Success: true,
	}, nil
}
