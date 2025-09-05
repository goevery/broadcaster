package handler

import (
	"context"
	"errors"
	"time"

	"github.com/juanpmarin/broadcaster/internal/broadcaster"
)

type JoinRequest struct {
	ChannelId string
}

type JoinResponse struct {
	SubscriptionId string    `json:"subscriptionId,omitempty"`
	Timestamp      time.Time `json:"timestamp"`
}

type JoinHandler struct {
	channelIdValidator   *ChannelIdValidator
	subscriptionRegistry broadcaster.Registry
}

func NewJoinHandler(
	channelIdValidator *ChannelIdValidator,
	subscriptionRegistry broadcaster.Registry,
) *JoinHandler {

	return &JoinHandler{
		channelIdValidator,
		subscriptionRegistry,
	}
}

func (h *JoinHandler) Handle(ctx context.Context, req JoinRequest) (JoinResponse, error) {
	err := h.channelIdValidator.Validate(req.ChannelId)
	if err != nil {
		return JoinResponse{}, err
	}

	connection, ok := broadcaster.ConnectionFromContext(ctx)
	if !ok {
		return JoinResponse{}, errors.New("connection not found in context")
	}

	if !connection.IsAuthorized(req.ChannelId) {
		return JoinResponse{},
			NewError(ErrorCodeUnauthenticated, errors.New("user not authorized to access this channel"))
	}

	err = h.subscriptionRegistry.Subscribe(req.ChannelId, connection)
	if err != nil {
		return JoinResponse{}, err
	}

	return JoinResponse{
		SubscriptionId: connection.Id,
		Timestamp:      time.Now(),
	}, nil
}
