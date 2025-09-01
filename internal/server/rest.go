package server

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/juanpmarin/broadcaster/internal/handler"
	"go.uber.org/zap"
)

type RESTServer struct {
	logger *zap.Logger

	pushHandler *handler.PushHandler
}

func NewRESTServer(
	logger *zap.Logger,
	pushHandler *handler.PushHandler,
) *RESTServer {
	return &RESTServer{
		logger,
		pushHandler,
	}
}

func (s *RESTServer) Register(router *mux.Router) {
	router.HandleFunc("/push", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			return
		}

		var pushRequest handler.PushRequest
		err := json.NewDecoder(r.Body).Decode(&pushRequest)
		if err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		pushResponse, err := s.pushHandler.Handle(r.Context(), pushRequest)
		if err != nil {
			s.logger.Error("failed to handle push request", zap.Error(err))
			http.Error(w, "failed to handle push request", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(pushResponse)
		if err != nil {
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
			return
		}
	}).Methods("POST", "OPTIONS")
}
