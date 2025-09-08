package server

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/goevery/broadcaster/internal/auth"
	"github.com/goevery/broadcaster/internal/handler"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type RESTServer struct {
	logger *zap.Logger

	publishHandler   *handler.PublishHandler
	authenticator *auth.Authenticator
}

func NewRESTServer(
	logger *zap.Logger,
	publishHandler *handler.PublishHandler,
	authenticator *auth.Authenticator,
) *RESTServer {
	return &RESTServer{
		logger,
		publishHandler,
		authenticator,
	}
}

func (s *RESTServer) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *RESTServer) authenticationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "missing authorization header", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		authentication, err := s.authenticator.AuthenticateAPIKey(tokenString)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		if !authentication.IsPublisher() {
			http.Error(w, "invalid token for publisher authentication", http.StatusUnauthorized)
			return
		}

		ctx := auth.WithAuthentication(r.Context(), authentication)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *RESTServer) Register(router *mux.Router) {
	publishRouter := router.Methods("POST", "OPTIONS").Subrouter()
	publishRouter.Use(s.corsMiddleware, s.authenticationMiddleware)
	publishRouter.HandleFunc("/publish", func(w http.ResponseWriter, r *http.Request) {
		var publishRequest handler.PublishRequest
		err := json.NewDecoder(r.Body).Decode(&publishRequest)
		if err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		publishResponse, err := s.publishHandler.Handle(r.Context(), publishRequest)
		if err != nil {
			s.logger.Error("failed to handle publish request", zap.Error(err))
			http.Error(w, "failed to handle publish request", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(publishResponse)
		if err != nil {
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
			return
		}
	})

	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}).Methods("GET")
}
