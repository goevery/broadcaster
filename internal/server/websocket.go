package server

import (
	"context"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/juanpmarin/broadcaster/internal/broadcaster"
	"github.com/juanpmarin/broadcaster/internal/protocol"
	"github.com/juanpmarin/broadcaster/internal/registry"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/sourcegraph/jsonrpc2"
	"go.uber.org/zap"
)

type WebSocketServer struct {
	logger            *zap.Logger
	upgrader          *websocket.Upgrader
	rpcHandlerFactory *RPCHandlerFactory
	registry          registry.Registry
}

func NewWebSocketServer(
	logger *zap.Logger,
	upgrader *websocket.Upgrader,
	rpcHandlerFactory *RPCHandlerFactory,
	registry registry.Registry,
) *WebSocketServer {
	return &WebSocketServer{
		logger,
		upgrader,
		rpcHandlerFactory,
		registry,
	}
}

func (s *WebSocketServer) Register(router *mux.Router) {
	router.HandleFunc("/websocket", func(w http.ResponseWriter, r *http.Request) {
		conn, err := s.upgrader.Upgrade(w, r, nil)
		if err != nil {
			s.logger.Warn("failed to upgrade to websocket", zap.Error(err))

			return
		}

		err = s.setupConnection(r, conn)
		if err != nil {
			s.logger.Error("failed to set up websocket connection", zap.Error(err))
			conn.Close()

			return
		}

		connectionId, err := registry.GenerateConnectionId()
		if err != nil {
			s.logger.Error("failed to generate connection ID", zap.Error(err))
			conn.Close()
			return
		}

		// Get client IP from request
		clientIp := r.RemoteAddr
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			clientIp = xff
		}

		s.logger.Info("websocket connection established",
			zap.String("connectionId", connectionId),
			zap.String("clientIp", clientIp))

		conn.SetReadLimit(1024)

		handlerLogger := s.logger.With(
			zap.String("connectionId", connectionId),
			zap.String("clientIp", clientIp))

		// Create handler first
		jsonrpcHandler := s.rpcHandlerFactory.New(handlerLogger)

		jsonrpcConn := jsonrpc2.NewConn(
			r.Context(),
			NewWebSocketObjectStream(conn),
			jsonrpcHandler,
			jsonrpc2.SetLogger(jsonrpcHandler),
		)

		// Create connection info with jsonrpc connection
		connInfo := registry.ConnectionInfo{
			Id:         connectionId,
			ClientIp:   clientIp,
			UserId:     "", // Can be set later based on authentication
			Connection: jsonrpcConn,
		}

		// Set connection info on handler
		jsonrpcHandler.SetConnectionInfo(connInfo)

		// Add connection info to context
		ctx := registry.WithConnectionInfo(r.Context(), connInfo)

		<-jsonrpcConn.DisconnectNotify()

		// Clean up subscriptions when connection closes
		err = s.registry.UnsubscribeAll(ctx, connectionId)
		if err != nil {
			s.logger.Error("failed to clean up subscriptions",
				zap.String("connectionId", connectionId),
				zap.Error(err))
		}

		s.logger.Info("websocket connection closed",
			zap.String("connectionId", connectionId))
	})

}

func (s *WebSocketServer) setupConnection(ctx context.Context, r *http.Request, conn *websocket.Conn) error {
	connectionId := gonanoid.Must()
	sendChan := make(chan broadcaster.Message, 1024)

	connection := broadcaster.Connection{
		Id:   connectionId,
		Send: sendChan,
	}

	ctx = broadcaster.WithConnection(ctx, connection)

	return nil
}

func (s *WebSocketServer) readPump(conn *websocket.Conn) {
	defer conn.Close()

	conn.SetReadLimit(1024)
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))

	for {
		var request protocol.Request
		err := conn.ReadJSON(&request)
		if err != nil {
			s.logger.Error("failed to read request", zap.Error(err))

			break
		}

		conn.SetReadDeadline(time.Now().Add(60 * time.Second))

		ctx := context.Background()
		response, err := s.handleRequest(ctx, request)
		if err != nil {
			s.logger.Error("failed to handle request", zap.Error(err))

			continue
		}

		err = conn.WriteJSON(response)
		if err != nil {
			s.logger.Error("failed to write response", zap.Error(err))

			break
		}
	}
}

func (s *WebSocketServer) handleRequest(ctx context.Context, request protocol.Request) (protocol.Response, error) {
	return protocol.Response{}, nil
}
