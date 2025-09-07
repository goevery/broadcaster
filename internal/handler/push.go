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

type PushRequest struct {
	ChannelId string `json:"channelId"`
	Payload   any    `json:"payload"`
}

type PushHandlerInterface interface {
	Handle(ctx context.Context, req PushRequest) (broadcaster.Message, error)
}

type PushHandler struct {
	channelIdValidator   *ChannelIdValidator
	subscriptionRegistry broadcaster.Registry
}

func NewPushHandler(
	channelIdValidator *ChannelIdValidator,
	subscriptionRegistry broadcaster.Registry,
) *PushHandler {
	return &PushHandler{
		channelIdValidator,
		subscriptionRegistry,
	}
}

func (h *PushHandler) Handle(ctx context.Context, req PushRequest) (broadcaster.Message, error) {
	authentication, ok := auth.AuthenticationFromContext(ctx)
	if !ok {
		return broadcaster.Message{}, errors.New("authentication not found in context")
	}

	if !authentication.IsAuthorized(req.ChannelId) {
		return broadcaster.Message{}, ierr.New(ierr.ErrorCodePermissionDenied, errors.New("not authorized to publish to this channel"))
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
