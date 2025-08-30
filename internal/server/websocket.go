package server

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
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

func (s *WebSocketServer) Register(mux *http.ServeMux) error {
	mux.HandleFunc("/websocket", func(w http.ResponseWriter, r *http.Request) {
		s.logger.Info("websocket endpoint hit")

		conn, err := s.upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}

		conn.WriteJSON("Hello!")
	})

	return nil
}
