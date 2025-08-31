package handler

import (
	"context"
	"errors"
	"regexp"
)

type PushRequest struct {
	ChannelId string `json:"channelId"`
	Payload   any    `json:"payload"`
}

type PushResponse struct {
}

type PushHandler struct {
	channelIdRegexp *regexp.Regexp
}

func NewPushHandler() *PushHandler {
	channelIdRegexp := regexp.MustCompile(ChannelIdRegex)

	return &PushHandler{
		channelIdRegexp,
	}
}

func (h *PushHandler) Handle(ctx context.Context, req PushRequest) (PushResponse, error) {
	validChannelId := h.channelIdRegexp.MatchString(req.ChannelId)
	if !validChannelId {
		return PushResponse{}, NewError(ErrorCodeInvalidArgument, errors.New("invalid channelId"))
	}

	return PushResponse{}, nil
}
