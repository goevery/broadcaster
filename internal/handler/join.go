package handler

import (
	"context"
	"errors"
	"time"

	"github.com/juanpmarin/broadcaster/internal/persistence"
	"github.com/juanpmarin/broadcaster/internal/protocol"
	"github.com/juanpmarin/broadcaster/internal/registry"
)

type JoinRequest struct {
	ChannelId         string
	LastSeenMessageId string
}

type JoinResponse struct {
	Timestamp      time.Time          `json:"timestamp"`
	History        []protocol.Message `json:"history"`
	SubscriptionId string             `json:"subscriptionId,omitempty"`
}

type JoinHandler struct {
	channelIdValidator   *ChannelIdValidator
	persistenceEngine    persistence.Engine
	subscriptionRegistry registry.Registry
}

func NewJoinHandler(
	channelIdValidator *ChannelIdValidator,
	persistenceEngine persistence.Engine,
	subscriptionRegistry registry.Registry,
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

	response := JoinResponse{
		Timestamp: time.Now(),
		History:   history,
	}

	// Subscribe the connection to the channel - connection info MUST be available
	connInfo, ok := registry.ConnectionInfoFromContext(ctx)
	if !ok {
		return JoinResponse{},
			protocol.NewError(protocol.ErrorCodeInvalidArgument, errors.New("connection info not available"))
	}

	subscription, err := h.subscriptionRegistry.Subscribe(ctx, req.ChannelId, connInfo)
	if err != nil {
		return JoinResponse{}, err
	}
	response.SubscriptionId = subscription.ConnectionInfo.Id

	return response, nil
}
