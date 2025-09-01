package server

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/juanpmarin/broadcaster/internal/handler"
	"go.uber.org/zap"
)

type Router struct {
	logger *zap.Logger

	heartbeatHandler *handler.HeartbeatHandler
	joinHandler      *handler.JoinHandler
	leaveHandler     *handler.LeaveHandler
	pushHandler      *handler.PushHandler
}

func NewRouter(
	logger *zap.Logger,
	heartbeatHandler *handler.HeartbeatHandler,
	joinHandler *handler.JoinHandler,
	leaveHandler *handler.LeaveHandler,
	pushHandler *handler.PushHandler,
) *Router {
	return &Router{
		logger,
		heartbeatHandler,
		joinHandler,
		leaveHandler,
		pushHandler,
	}
}

func (r *Router) RouteRequest(ctx context.Context, request handler.Request) *handler.Response {
	response, err := r.Handle(ctx, request)
	if err != nil {
		response := request.ReplyWithError(r.mapError(err))

		return &response
	}

	hasResponse := response != nil

	if request.ReplyExpected() && !hasResponse {
		r.logger.Error("handler did not return a response but one was expected", zap.String("method", request.Method))

		response := request.ReplyWithError(
			handler.NewError(handler.ErrorCodeInternal, errors.New("internal error")),
		)

		return &response
	}

	if !request.ReplyExpected() && hasResponse {
		r.logger.Error("handler returned a response but none was expected", zap.String("method", request.Method))

		return nil
	}

	if hasResponse {
		rawJson, err := json.Marshal(response)
		if err != nil {
			response := request.ReplyWithError(r.mapError(err))

			return &response
		}

		payload := json.RawMessage(rawJson)
		response := request.Reply(&payload)

		return &response
	}

	return nil
}

func (r *Router) Handle(ctx context.Context, request handler.Request) (any, error) {
	switch request.Method {
	case "heartbeat":
		return r.heartbeatHandler.Handle(), nil
	case "join":
		var joinReq handler.JoinRequest
		if err := decodeParams(request.Params, &joinReq); err != nil {
			return nil, err
		}

		return r.joinHandler.Handle(ctx, joinReq)
	case "leave":
		var leaveReq handler.LeaveRequest
		if err := decodeParams(request.Params, &leaveReq); err != nil {
			return nil, err
		}

		return r.leaveHandler.Handle(ctx, leaveReq)
	case "push":
		var pushReq handler.PushRequest
		if err := decodeParams(request.Params, &pushReq); err != nil {
			return nil, err
		}

		return r.pushHandler.Handle(ctx, pushReq)
	default:
		return nil, handler.NewError(handler.ErrorCodeNotFound, errors.New("method not found: "+request.Method))
	}
}

func (r *Router) mapError(err error) handler.Error {
	var handlerErr handler.Error
	if errors.As(err, &handlerErr) {
		return handlerErr
	}

	r.logger.Error("error in rpc handler", zap.Error(err))

	return handler.NewError(handler.ErrorCodeInternal, errors.New("internal error"))
}

func decodeParams(params *json.RawMessage, v any) error {
	if params == nil {
		return handler.NewError(handler.ErrorCodeInvalidArgument, errors.New("missing params"))
	}

	if err := json.Unmarshal(*params, v); err != nil {
		return handler.NewError(handler.ErrorCodeInvalidArgument, errors.New("invalid params: "+err.Error()))
	}

	return nil
}
