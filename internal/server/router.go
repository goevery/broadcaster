package server

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/goevery/broadcaster/internal/handler"
	"github.com/goevery/broadcaster/internal/ierr"
	"go.uber.org/zap"
)

type Router struct {
	logger *zap.Logger

	heartbeatHandler handler.HeartbeatHandlerInterface
	joinHandler      handler.JoinHandlerInterface
	leaveHandler     handler.LeaveHandlerInterface
	pushHandler      handler.PushHandlerInterface
	authHandler      handler.AuthHandlerInterface
}

func NewRouter(
	logger *zap.Logger,
	heartbeatHandler handler.HeartbeatHandlerInterface,
	joinHandler handler.JoinHandlerInterface,
	leaveHandler handler.LeaveHandlerInterface,
	pushHandler handler.PushHandlerInterface,
	authHandler handler.AuthHandlerInterface,
) *Router {
	return &Router{
		logger,
		heartbeatHandler,
		joinHandler,
		leaveHandler,
		pushHandler,
		authHandler,
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
			ierr.New(ierr.ErrorCodeInternal, errors.New("internal error")),
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
	case "auth":
		var authReq handler.AuthRequest
		if err := decodeParams(request.Params, &authReq); err != nil {
			return nil, err
		}
		return r.authHandler.Handle(ctx, authReq)
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
		return nil, ierr.New(ierr.ErrorCodeNotFound, errors.New("method not found: "+request.Method))
	}
}

func (r *Router) mapError(err error) ierr.Error {
	var handlerErr ierr.Error
	if errors.As(err, &handlerErr) {
		return handlerErr
	}

	r.logger.Error("error in rpc handler", zap.Error(err))

	return ierr.New(ierr.ErrorCodeInternal, errors.New("internal error"))
}

func decodeParams(params *json.RawMessage, v any) error {
	if params == nil {
		return ierr.New(ierr.ErrorCodeInvalidArgument, errors.New("missing params"))
	}

	if err := json.Unmarshal(*params, v); err != nil {
		return ierr.New(ierr.ErrorCodeInvalidArgument, errors.New("invalid params: "+err.Error()))
	}

	return nil
}
