package handler

import (
	"context"
	"errors"

	"github.com/juanpmarin/broadcaster/internal/protocol"
	"github.com/juanpmarin/broadcaster/internal/registry"
)

type LeaveRequest struct {
	ChannelId string `json:"channelId"`
}

type LeaveResponse struct {
	Success bool `json:"success"`
}

type LeaveHandler struct {
	channelIdValidator   *ChannelIdValidator
	subscriptionRegistry registry.Registry
}

func NewLeaveHandler(
	channelIdValidator *ChannelIdValidator,
	subscriptionRegistry registry.Registry,
) *LeaveHandler {
	return &LeaveHandler{
		channelIdValidator,
		subscriptionRegistry,
	}
}

func (h *LeaveHandler) Handle(ctx context.Context, req LeaveRequest) (LeaveResponse, error) {
	err := h.channelIdValidator.Validate(req.ChannelId)
	if err != nil {
		return LeaveResponse{},
			protocol.NewError(protocol.ErrorCodeInvalidArgument, errors.New("invalid channelId"))
	}

	// Connection info MUST be available for leave operation
	connInfo, ok := registry.ConnectionInfoFromContext(ctx)
	if !ok {
		return LeaveResponse{},
			protocol.NewError(protocol.ErrorCodeInvalidArgument, errors.New("connection info not available"))
	}

	// Unsubscribe the connection from the channel
	err = h.subscriptionRegistry.Unsubscribe(ctx, req.ChannelId, connInfo.Id)
	if err != nil {
		return LeaveResponse{}, err
	}

	return LeaveResponse{
		Success: true,
	}, nil
}
