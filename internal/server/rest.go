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

	pushHandler   *handler.PushHandler
	authenticator *auth.Authenticator
}

func NewRESTServer(
	logger *zap.Logger,
	pushHandler *handler.PushHandler,
	authenticator *auth.Authenticator,
) *RESTServer {
	return &RESTServer{
		logger,
		pushHandler,
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
	pushRouter := router.Methods("POST", "OPTIONS").Subrouter()
	pushRouter.Use(s.corsMiddleware, s.authenticationMiddleware)
	pushRouter.HandleFunc("/push", func(w http.ResponseWriter, r *http.Request) {
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
	})

	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}).Methods("GET")
}
