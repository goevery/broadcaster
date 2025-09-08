package handler

import (
	"context"
	"errors"
	"time"

	"github.com/goevery/broadcaster/internal/broadcaster"
	"github.com/goevery/broadcaster/internal/ierr"
)

type SubscribeRequest struct {
	Channel string `json:"channel"`
}

type SubscribeResponse struct {
	SubscriptionId string    `json:"subscriptionId,omitempty"`
	Timestamp      time.Time `json:"timestamp"`
}

type SubscribeHandlerInterface interface {
	Handle(ctx context.Context, req SubscribeRequest) (SubscribeResponse, error)
}

type SubscribeHandler struct {
	channelValidator     *ChannelValidator
	subscriptionRegistry broadcaster.Registry
}

func NewSubscribeHandler(
	channelValidator *ChannelValidator,
	subscriptionRegistry broadcaster.Registry,
) *SubscribeHandler {

	return &SubscribeHandler{
		channelValidator,
		subscriptionRegistry,
	}
}

func (h *SubscribeHandler) Handle(ctx context.Context, req SubscribeRequest) (SubscribeResponse, error) {
	err := h.channelValidator.Validate(req.Channel)
	if err != nil {
		return SubscribeResponse{}, err
	}

	connection, ok := broadcaster.ConnectionFromContext(ctx)
	if !ok {
		return SubscribeResponse{}, errors.New("connection not found in context")
	}

	auth := connection.GetAuthentication()
	if auth == nil {
		return SubscribeResponse{},
			ierr.New(ierr.ErrorCodeUnauthenticated, errors.New("authentication required"))
	}

	if !auth.IsSubscriber() {
		return SubscribeResponse{},
			ierr.New(ierr.ErrorCodePermissionDenied, errors.New("subscribe scope required to subscribe to a channel"))
	}

	if !connection.IsAuthorized(req.Channel) {
		return SubscribeResponse{},
			ierr.New(ierr.ErrorCodeUnauthenticated, errors.New("user not authorized to access this channel"))
	}

	err = h.subscriptionRegistry.Subscribe(req.Channel, connection.Id)
	if err != nil {
		return SubscribeResponse{}, err
	}

	return SubscribeResponse{
		SubscriptionId: connection.Id,
		Timestamp:      time.Now(),
	}, nil
}
