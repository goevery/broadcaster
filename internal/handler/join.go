package handler

import (
	"context"
	"errors"
	"regexp"
)

const ChannelIdRegex = `^(\w+:?)+$`

type JoinRequest struct {
	ChannelId string `json:"channelId"`
}

type JoinResponse struct {
}

type JoinHandler struct {
	channelIdRegexp *regexp.Regexp
}

func NewJoinHandler() *JoinHandler {
	channelIdRegexp := regexp.MustCompile(ChannelIdRegex)

	return &JoinHandler{
		channelIdRegexp,
	}
}

func (h *JoinHandler) Handle(ctx context.Context, req JoinRequest) (JoinResponse, error) {
	validChannelId := h.channelIdRegexp.MatchString(req.ChannelId)
	if !validChannelId {
		return JoinResponse{}, NewError(ErrorCodeInvalidArgument, errors.New("invalid channelId"))
	}

	return JoinResponse{}, nil
}
