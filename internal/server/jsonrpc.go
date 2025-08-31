package server

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/juanpmarin/broadcaster/internal/handler"
	"github.com/sourcegraph/jsonrpc2"
	"go.uber.org/zap"
)

type RPCHandlerFactory struct {
	heartbeatHandler *handler.HeartbeatHandler
	joinHandler      *handler.JoinHandler
	pushHandler      *handler.PushHandler
}

func NewRPCHandlerFactory(
	heartbeatHandler *handler.HeartbeatHandler,
	joinHandler *handler.JoinHandler,
	pushHandler *handler.PushHandler,
) *RPCHandlerFactory {
	return &RPCHandlerFactory{
		heartbeatHandler,
		joinHandler,
		pushHandler,
	}
}

func (f *RPCHandlerFactory) New(logger *zap.Logger) *RPCHandler {
	return &RPCHandler{
		logger,
		logger.Sugar(),

		f.heartbeatHandler,
		f.joinHandler,
		f.pushHandler,
	}
}

type RPCHandler struct {
	logger        *zap.Logger
	sugaredLogger *zap.SugaredLogger

	heartbeatHandler *handler.HeartbeatHandler
	joinHandler      *handler.JoinHandler
	pushHandler      *handler.PushHandler
}

func (h *RPCHandler) Handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) {
	defer func() {
		if r := recover(); r != nil {
			panicErr, ok := r.(error)
			if ok {
				h.logger.Error("panic in rpc handler", zap.Error(panicErr))
			} else {
				h.logger.Error("panic in rpc handler", zap.Any("error", r))
			}

			_ = conn.Close()
		}
	}()

	h.logger.Debug("jsonrpc2 request received",
		zap.String("method", req.Method),
		zap.Any("params", req.Params),
		zap.String("id", req.ID.String()))

	timeoutCtx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()

	response, err := h.routeRequest(timeoutCtx, req)
	if err != nil {
		h.logger.Error("error in rpc handler", zap.Error(err))

		jsonrpc2Err, ok := err.(*jsonrpc2.Error)
		if !ok {
			jsonrpc2Err = &jsonrpc2.Error{
				Code:    jsonrpc2.CodeInternalError,
				Message: "internal server error",
			}
		}

		err = conn.ReplyWithError(ctx, req.ID, jsonrpc2Err)
		if err != nil {
			h.logger.Error("failed to reply", zap.Error(err))
		}

		return
	}

	hasResponse := response != nil
	responseExpected := req.Notif == false

	if responseExpected != hasResponse {
		h.logger.Error("response expected mismatch",
			zap.Bool("responseExpected", responseExpected),
			zap.Bool("hasResponse", hasResponse),
		)

		err = conn.ReplyWithError(ctx, req.ID, &jsonrpc2.Error{
			Code:    jsonrpc2.CodeInternalError,
			Message: "internal server error",
		})
		if err != nil {
			h.logger.Error("failed to reply", zap.Error(err))
		}

		return
	}

	if hasResponse {
		err = conn.Reply(ctx, req.ID, response)
		if err != nil {
			h.logger.Error("failed to reply", zap.Error(err))
		}
	}
}

func (h *RPCHandler) Printf(format string, v ...any) {
	h.sugaredLogger.Infof(strings.TrimSpace(format), v)
}

func (h *RPCHandler) routeRequest(ctx context.Context, req *jsonrpc2.Request) (any, error) {
	if req.Method == "heartbeat" {
		return h.heartbeatHandler.Handle(), nil
	}

	if req.Method == "join" {
		var joinReq handler.JoinRequest
		err := h.decodeParams(req.Params, &joinReq)
		if err != nil {
			return nil, err
		}

		return h.joinHandler.Handle(ctx, joinReq)
	}

	if req.Method == "push" {
		var pushReq handler.PushRequest
		err := h.decodeParams(req.Params, &pushReq)
		if err != nil {
			return nil, err
		}

		return h.pushHandler.Handle(ctx, pushReq)
	}

	return nil, &jsonrpc2.Error{
		Code:    jsonrpc2.CodeMethodNotFound,
		Message: "method not found",
	}
}

func (h *RPCHandler) decodeParams(params *json.RawMessage, v any) error {
	if params == nil {
		return &jsonrpc2.Error{
			Code:    jsonrpc2.CodeInvalidParams,
			Message: "missing params",
		}
	}

	return json.Unmarshal(*params, v)
}

func (h *RPCHandler) mapError(err error) *jsonrpc2.Error {
	if err == nil {
		return nil
	}

	jsonrpc2Err, ok := err.(*jsonrpc2.Error)
	if ok {
		return jsonrpc2Err
	}

	handlerErr, ok := err.(*handler.Error)
	if ok {
		return &jsonrpc2.Error{
			Code:    mapErrorCode(handlerErr.Code()),
			Message: handlerErr.Error(),
		}
	}

	h.logger.Error("error in rpc handler", zap.Error(err))

	return &jsonrpc2.Error{
		Code:    jsonrpc2.CodeInternalError,
		Message: "internal server error",
	}
}

func mapErrorCode(code handler.ErrorCode) int64 {
	switch code {
	case handler.ErrorCodeInvalidArgument:
		return jsonrpc2.CodeInvalidParams
	default:
		return jsonrpc2.CodeInternalError
	}
}
