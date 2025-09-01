package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/juanpmarin/broadcaster/internal/broadcaster"
	"github.com/juanpmarin/broadcaster/internal/handler"
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
		sendChan := make(chan broadcaster.Message, 1024)

		broadcasterConn := broadcaster.Connection{
			Id:   connectionId,
			Send: sendChan,
		}

		ctx := broadcaster.WithConnection(r.Context(), broadcasterConn)

		go s.readPump(ctx, wsConn, broadcasterConn)
		go s.writePump(ctx, wsConn, broadcasterConn)
	})

}

func (s *WebSocketServer) readPump(
	ctx context.Context,
	wsConn *websocket.Conn,
	broadcasterConn broadcaster.Connection,
) {
	defer func() {
		s.registry.Disconnect(broadcasterConn.Id)
	}()

	wsConn.SetReadLimit(1024)
	wsConn.SetReadDeadline(time.Now().Add(60 * time.Second))

	for {
		var request handler.Request
		err := wsConn.ReadJSON(&request)
		if err != nil {
			s.logger.Error("failed to read request", zap.Error(err))

			break
		}

		wsConn.SetReadDeadline(time.Now().Add(60 * time.Second))

		response := s.router.RouteRequest(ctx, request)

		err = wsConn.WriteJSON(response)
		if err != nil {
			s.logger.Error("failed to write response", zap.Error(err))

			break
		}
	}
}

func (s *WebSocketServer) writePump(
	ctx context.Context,
	wsConn *websocket.Conn,
	broadcasterConn broadcaster.Connection,
) {
	defer func() {
		_ = wsConn.Close()
	}()

	for {
		select {
		case message, ok := <-broadcasterConn.Send:
			if !ok {
				// Channel closed, close the connection
				wsConn.WriteMessage(websocket.CloseMessage, []byte{})

				return
			}

			err := sendBroadcastNotification(wsConn, message)
			if err != nil {
				s.logger.Error("failed to send broadcast notification", zap.Error(err))

				return
			}
		case <-ctx.Done():
			return
		}
	}
}

func sendBroadcastNotification(conn *websocket.Conn, message broadcaster.Message) error {
	rawJson, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	payload := json.RawMessage(rawJson)
	notification := handler.NewNotification("broadcast", &payload)

	err = conn.WriteJSON(notification)
	if err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	return nil
}
