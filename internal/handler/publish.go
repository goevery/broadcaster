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
	ChannelId string `json:"channelId"`
	Payload   any    `json:"payload"`
}

type PublishHandlerInterface interface {
	Handle(ctx context.Context, req PublishRequest) (broadcaster.Message, error)
}

type PublishHandler struct {
	channelIdValidator   *ChannelIdValidator
	subscriptionRegistry broadcaster.Registry
}

func NewPublishHandler(
	channelIdValidator *ChannelIdValidator,
	subscriptionRegistry broadcaster.Registry,
) *PublishHandler {
	return &PublishHandler{
		channelIdValidator,
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

	if !authentication.IsAuthorized(req.ChannelId) {
		return broadcaster.Message{},
			ierr.New(ierr.ErrorCodePermissionDenied, errors.New("user not authorized to publish to this channel"))
	}

	err := h.channelIdValidator.Validate(req.ChannelId)
	if err != nil {
		return broadcaster.Message{}, err
	}

	message := broadcaster.Message{
		Id:         gonanoid.Must(),
		CreateTime: time.Now(),
		ChannelId:  req.ChannelId,
		Payload:    req.Payload,
	}

	h.subscriptionRegistry.Broadcast(message)

	return message, nil
}
