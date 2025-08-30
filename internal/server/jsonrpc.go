package server

import (
	"context"
	"strings"
	"time"

	"github.com/sourcegraph/jsonrpc2"
	"go.uber.org/zap"
)

type RPCHandler struct {
	logger        *zap.Logger
	sugaredLogger *zap.SugaredLogger
}

func NewRPCHandler(logger *zap.Logger) *RPCHandler {
	return &RPCHandler{
		logger,
		logger.Sugar(),
	}
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

		err = conn.ReplyWithError(ctx, req.ID, &jsonrpc2.Error{
			Code:    jsonrpc2.CodeInternalError,
			Message: "internal server error",
		})
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
	return nil, nil
}
