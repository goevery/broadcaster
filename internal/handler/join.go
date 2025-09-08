package handler

import (
	"context"
	"errors"
	"time"

	"github.com/goevery/broadcaster/internal/broadcaster"
	"github.com/goevery/broadcaster/internal/ierr"
)

type JoinRequest struct {
	ChannelId string
}

type JoinResponse struct {
	SubscriptionId string    `json:"subscriptionId,omitempty"`
	Timestamp      time.Time `json:"timestamp"`
}

type JoinHandlerInterface interface {
	Handle(ctx context.Context, req JoinRequest) (JoinResponse, error)
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

	auth := connection.GetAuthentication()
	if auth == nil {
		return JoinResponse{},
			ierr.New(ierr.ErrorCodeUnauthenticated, errors.New("authentication required"))
	}

	if !auth.IsSubscriber() {
		return JoinResponse{},
			ierr.New(ierr.ErrorCodePermissionDenied, errors.New("subscribe scope required to join a channel"))
	}

	if !connection.IsAuthorized(req.ChannelId) {
		return JoinResponse{},
			ierr.New(ierr.ErrorCodeUnauthenticated, errors.New("user not authorized to access this channel"))
	}

	err = h.subscriptionRegistry.Subscribe(req.ChannelId, connection.Id)
	if err != nil {
		return JoinResponse{}, err
	}

	return JoinResponse{
		SubscriptionId: connection.Id,
		Timestamp:      time.Now(),
	}, nil
}
