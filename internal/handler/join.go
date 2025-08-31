package handler

import (
	"context"
	"errors"
	"time"

	"github.com/juanpmarin/broadcaster/internal/persistence"
	"github.com/juanpmarin/broadcaster/internal/protocol"
)

type JoinRequest struct {
	ChannelId         string
	LastSeenMessageId string
}

type JoinResponse struct {
	Timestamp time.Time          `json:"timestamp"`
	History   []protocol.Message `json:"history"`
}

type JoinHandler struct {
	channelIdValidator *ChannelIdValidator
	persistenceEngine  persistence.Engine
}

func NewJoinHandler(
	channelIdValidator *ChannelIdValidator,
	persistenceEngine persistence.Engine,
) *JoinHandler {

	return &JoinHandler{
		channelIdValidator,
		persistenceEngine,
	}
}

func (h *JoinHandler) Handle(ctx context.Context, req JoinRequest) (JoinResponse, error) {
	err := h.channelIdValidator.Validate(req.ChannelId)
	if err != nil {
		return JoinResponse{},
			protocol.NewError(protocol.ErrorCodeInvalidArgument, errors.New("invalid channelId"))
	}

	var history []protocol.Message

	if req.LastSeenMessageId != "" {
		history, err = h.persistenceEngine.List(ctx, req.ChannelId, req.LastSeenMessageId)
		if err != nil {
			return JoinResponse{}, err
		}
	}

	return JoinResponse{
		Timestamp: time.Now(),
		History:   history,
	}, nil
}
