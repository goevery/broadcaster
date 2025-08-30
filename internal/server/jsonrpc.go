package server

import (
	"context"
	"strings"

	"github.com/sourcegraph/jsonrpc2"
	"go.uber.org/zap"
)

type RPCHandler struct {
	logger *zap.Logger
}

func NewRPCHandler(logger *zap.Logger) *RPCHandler {
	return &RPCHandler{
		logger,
	}
}

func (h *RPCHandler) Handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) {
	h.logger.Debug("jsonrpc2 request received",
		zap.String("method", req.Method),
		zap.Any("params", req.Params),
		zap.String("params", req.ID.String()))

	switch req.Method {
	case "heartbeat":
		conn.Reply(ctx, req.ID, "ok")
	default:
		h.logger.Warn("unknown method", zap.String("method", req.Method))
		if err := conn.ReplyWithError(ctx, req.ID, &jsonrpc2.Error{
			Code:    jsonrpc2.CodeMethodNotFound,
			Message: "method not found",
		}); err != nil {
			h.logger.Error("failed to reply with error", zap.Error(err))
		}
	}
}

type RPCLogger struct {
	logger *zap.SugaredLogger
}

func NewRPCLogger(logger *zap.Logger) *RPCLogger {
	return &RPCLogger{
		logger.Sugar(),
	}
}

func (l *RPCLogger) Printf(format string, v ...any) {
	l.logger.Infof(strings.TrimSpace(format), v)
}
