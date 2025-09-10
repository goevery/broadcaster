package server

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/goevery/broadcaster/internal/broadcaster"
	"github.com/goevery/broadcaster/internal/handler"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"go.uber.org/zap"
)

type WebSocketServer struct {
	logger   *zap.Logger
	upgrader *websocket.Upgrader
	registry broadcaster.Registry
	router   *Router
}

func NewWebSocketServer(
	logger *zap.Logger,
	upgrader *websocket.Upgrader,
	registry broadcaster.Registry,
	router *Router,
) *WebSocketServer {
	return &WebSocketServer{
		logger,
		upgrader,
		registry,
		router,
	}
}

func (s *WebSocketServer) Register(router *mux.Router) {
	router.HandleFunc("/websocket", func(w http.ResponseWriter, r *http.Request) {
		wsConn, err := s.upgrader.Upgrade(w, r, nil)
		if err != nil {
			s.logger.Warn("failed to upgrade to websocket", zap.Error(err))

			return
		}

		connectionId := gonanoid.Must()
		broascasterChannel := make(chan broadcaster.Message, 1024)

		rpcChannel := make(chan any, 1024)

		broadcasterConn := &broadcaster.Connection{
			Id:   connectionId,
			Send: broascasterChannel,
			Seq:  0,
		}

		s.registry.Connect(broadcasterConn)

		ctx := broadcaster.WithConnection(r.Context(), broadcasterConn)

		go s.readPump(ctx, wsConn, rpcChannel, connectionId)
		go s.writePump(ctx, wsConn, rpcChannel)

		for message := range broascasterChannel {
			rawJson, err := json.Marshal(message)
			if err != nil {
				s.logger.Error("failed to marshal message", zap.Error(err))

				return
			}

			payload := json.RawMessage(rawJson)
			notification := handler.NewNotification("broadcast", &payload)

			rpcChannel <- notification
		}

		close(rpcChannel)

		s.logger.Info("websocket connection closed", zap.String("connectionId", connectionId))
	})
}

func (s *WebSocketServer) readPump(
	ctx context.Context,
	wsConn *websocket.Conn,
	rpcChannel chan any,
	connectionId string,
) {
	defer func() {
		s.registry.Disconnect(connectionId)
	}()

	wsConn.SetReadLimit(1024)
	wsConn.SetReadDeadline(time.Now().Add(60 * time.Second))

	for {
		var request handler.Request
		err := wsConn.ReadJSON(&request)
		if err != nil {
			isExpectedClose := websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure)
			if !isExpectedClose {
				s.logger.Error("failed to read request", zap.Error(err))
			}

			break
		}

		wsConn.SetReadDeadline(time.Now().Add(60 * time.Second))

		response := s.router.RouteRequest(ctx, request)
		if response != nil {
			rpcChannel <- response
		}
	}
}

func (s *WebSocketServer) writePump(
	ctx context.Context,
	wsConn *websocket.Conn,
	rpcChannel chan any,
) {
	defer func() {
		_ = wsConn.Close()
	}()

	for {
		select {
		case message, ok := <-rpcChannel:
			if !ok {
				// Channel closed, close the connection
				wsConn.SetWriteDeadline(time.Now().Add(10 * time.Second))
				wsConn.WriteMessage(websocket.CloseMessage, []byte{})

				return
			}

			wsConn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			err := wsConn.WriteJSON(message)
			if err != nil {
				s.logger.Error("failed to send broadcast notification", zap.Error(err))

				return
			}
		case <-ctx.Done():
			return
		}
	}
}
