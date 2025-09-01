package handler

import (
	"context"
	"errors"

	"github.com/juanpmarin/broadcaster/internal/broadcaster"
)

type LeaveRequest struct {
	ChannelId string `json:"channelId"`
}

type LeaveResponse struct {
	Success bool `json:"success"`
}

type LeaveHandler struct {
	channelIdValidator   *ChannelIdValidator
	subscriptionRegistry broadcaster.Registry
}

func NewLeaveHandler(
	channelIdValidator *ChannelIdValidator,
	subscriptionRegistry broadcaster.Registry,
) *LeaveHandler {
	return &LeaveHandler{
		channelIdValidator,
		subscriptionRegistry,
	}
}

func (h *LeaveHandler) Handle(ctx context.Context, req LeaveRequest) (LeaveResponse, error) {
	err := h.channelIdValidator.Validate(req.ChannelId)
	if err != nil {
		return LeaveResponse{}, err
	}

	connection, ok := broadcaster.ConnectionFromContext(ctx)
	if !ok {
		return LeaveResponse{},
			NewError(ErrorCodeInvalidArgument, errors.New("connection info not available"))
	}

	h.subscriptionRegistry.Unregister(req.ChannelId, connection.Id)

	return LeaveResponse{
		Success: true,
	}, nil
}
