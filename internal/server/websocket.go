package server

import (
	"context"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/sourcegraph/jsonrpc2"
	"go.uber.org/zap"
)

type WebSocketServer struct {
	logger   *zap.Logger
	upgrader *websocket.Upgrader
}

func NewWebSocketServer(
	logger *zap.Logger,
	upgrader *websocket.Upgrader,
) *WebSocketServer {
	return &WebSocketServer{
		logger,
		upgrader,
	}
}

func (s *WebSocketServer) Register(ctx context.Context, mux *http.ServeMux) error {
	mux.HandleFunc("/websocket", func(w http.ResponseWriter, r *http.Request) {
		s.logger.Info("websocket endpoint hit")

		conn, err := s.upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}

		s.logger.Info("websocket connection established")

		conn.SetReadLimit(1024)

		jsonrpcHandler := NewRPCHandler(s.logger.With(zap.String("clientIp", "FIX-ME-CLIENT-IP")))

		jsonrpcConn := jsonrpc2.NewConn(
			ctx,
			NewWebSocketObjectStream(conn),
			jsonrpcHandler,
			jsonrpc2.SetLogger(jsonrpcHandler),
		)

		<-jsonrpcConn.DisconnectNotify()

		s.logger.Info("websocket connection closed")
	})

	return nil
}

type WebSocketObjectStream struct {
	connection *websocket.Conn
}

func NewWebSocketObjectStream(connection *websocket.Conn) *WebSocketObjectStream {
	return &WebSocketObjectStream{
		connection,
	}
}

func (s *WebSocketObjectStream) WriteObject(obj any) error {
	return s.connection.WriteJSON(obj)
}

func (s *WebSocketObjectStream) ReadObject(v any) error {
	return s.connection.ReadJSON(v)
}

func (s *WebSocketObjectStream) Close() error {
	return s.connection.Close()
}
