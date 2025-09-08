package handler

import (
	"context"
	"errors"

	"github.com/goevery/broadcaster/internal/broadcaster"
)

type UnsubscribeRequest struct {
	ChannelId string `json:"channelId"`
}

type UnsubscribeResponse struct {
	Success bool `json:"success"`
}

type UnsubscribeHandlerInterface interface {
	Handle(ctx context.Context, req UnsubscribeRequest) (UnsubscribeResponse, error)
}

type UnsubscribeHandler struct {
	channelIdValidator   *ChannelIdValidator
	subscriptionRegistry broadcaster.Registry
}

func NewUnsubscribeHandler(
	channelIdValidator *ChannelIdValidator,
	subscriptionRegistry broadcaster.Registry,
) *UnsubscribeHandler {
	return &UnsubscribeHandler{
		channelIdValidator,
		subscriptionRegistry,
	}
}

func (h *UnsubscribeHandler) Handle(ctx context.Context, req UnsubscribeRequest) (UnsubscribeResponse, error) {
	err := h.channelIdValidator.Validate(req.ChannelId)
	if err != nil {
		return UnsubscribeResponse{}, err
	}

	connection, ok := broadcaster.ConnectionFromContext(ctx)
	if !ok {
		return UnsubscribeResponse{}, errors.New("connection not found in context")
	}

	h.subscriptionRegistry.Unsubscribe(req.ChannelId, connection.Id)

	return UnsubscribeResponse{
		Success: true,
	}, nil
}
