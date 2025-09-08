package handler

import (
	"context"
	"errors"
	"time"

	"github.com/goevery/broadcaster/internal/auth"
	"github.com/goevery/broadcaster/internal/broadcaster"
	"github.com/goevery/broadcaster/internal/ierr"
	gonanoid "github.com/matoous/go-nanoid/v2"
)

type PublishRequest struct {
	Channel string `json:"channel"`
	Event   string `json:"event"`
	Payload any    `json:"payload"`
}

type PublishHandlerInterface interface {
	Handle(ctx context.Context, req PublishRequest) (broadcaster.Message, error)
}

type PublishHandler struct {
	channelValidator     *ChannelValidator
	subscriptionRegistry broadcaster.Registry
}

func NewPublishHandler(
	channelValidator *ChannelValidator,
	subscriptionRegistry broadcaster.Registry,
) *PublishHandler {
	return &PublishHandler{
		channelValidator,
		subscriptionRegistry,
	}
}

func (h *PublishHandler) Handle(ctx context.Context, req PublishRequest) (broadcaster.Message, error) {
	var authentication *auth.Authentication

	connection, ok := broadcaster.ConnectionFromContext(ctx)
	if ok {
		authentication = connection.GetAuthentication()
	}

	if authentication == nil {
		authentication, ok = auth.AuthenticationFromContext(ctx)
		if !ok {
			return broadcaster.Message{}, ierr.New(ierr.ErrorCodeUnauthenticated, errors.New("user not authenticated"))
		}
	}

	if !authentication.IsPublisher() {
		return broadcaster.Message{},
			ierr.New(ierr.ErrorCodePermissionDenied, errors.New("user not authorized to publish messages"))
	}

	if !authentication.IsAuthorized(req.Channel) {
		return broadcaster.Message{},
			ierr.New(ierr.ErrorCodePermissionDenied, errors.New("user not authorized to publish to this channel"))
	}

	err := h.channelValidator.Validate(req.Channel)
	if err != nil {
		return broadcaster.Message{}, err
	}

	message := broadcaster.Message{
		Id:         gonanoid.Must(),
		CreateTime: time.Now(),
		Channel:    req.Channel,
		Event:      req.Event,
		Payload:    req.Payload,
	}

	h.subscriptionRegistry.Broadcast(message)

	return message, nil
}
