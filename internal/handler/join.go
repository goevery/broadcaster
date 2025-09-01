package handler

import (
	"context"
	"errors"
	"time"

	"github.com/juanpmarin/broadcaster/internal/broadcaster"
	"github.com/juanpmarin/broadcaster/internal/persistence"
)

type JoinRequest struct {
	ChannelId         string
	LastSeenMessageId string
}

type JoinResponse struct {
	SubscriptionId   string                `json:"subscriptionId,omitempty"`
	Timestamp        time.Time             `json:"timestamp"`
	History          []broadcaster.Message `json:"history,omitempty"`
	HistoryRecovered bool                  `json:"historyRecovered"`
}

type JoinHandler struct {
	channelIdValidator   *ChannelIdValidator
	persistenceEngine    persistence.Engine
	subscriptionRegistry broadcaster.Registry
}

func NewJoinHandler(
	channelIdValidator *ChannelIdValidator,
	persistenceEngine persistence.Engine,
	subscriptionRegistry broadcaster.Registry,
) *JoinHandler {

	return &JoinHandler{
		channelIdValidator,
		persistenceEngine,
		subscriptionRegistry,
	}
}

func (h *JoinHandler) Handle(ctx context.Context, req JoinRequest) (JoinResponse, error) {
	err := h.channelIdValidator.Validate(req.ChannelId)
	if err != nil {
		return JoinResponse{}, err
	}

	var history []broadcaster.Message
	historyRecovered := false

	if req.LastSeenMessageId != "" {
		history, err = h.persistenceEngine.List(ctx, req.ChannelId, req.LastSeenMessageId)
		if err != nil {
			return JoinResponse{}, err
		}

		for _, msg := range history {
			if msg.Id == req.LastSeenMessageId {
				historyRecovered = true
				break
			}
		}
	}

	connection, ok := broadcaster.ConnectionFromContext(ctx)
	if !ok {
		return JoinResponse{},
			NewError(ErrorCodeFailedPrecondition, errors.New("connection info not available"))
	}

	err = h.subscriptionRegistry.Subscribe(req.ChannelId, connection)
	if err != nil {
		return JoinResponse{}, err
	}

	return JoinResponse{
		SubscriptionId:   connection.Id,
		Timestamp:        time.Now(),
		History:          history,
		HistoryRecovered: historyRecovered,
	}, nil
}
