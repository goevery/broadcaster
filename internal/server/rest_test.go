package server

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/goevery/broadcaster/internal/auth"
	"github.com/goevery/broadcaster/internal/broadcaster"
	"github.com/goevery/broadcaster/internal/handler"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestRESTServer_Publish(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	authenticator := auth.NewAuthenticator("test-secret", []string{"test-api-key"})
	registry := broadcaster.NewMockRegistry(t)
	channelValidator := handler.NewChannelValidator()
	publishHandler := handler.NewPublishHandler(channelValidator, registry)

	restServer := NewRESTServer(logger, publishHandler, authenticator)

	router := mux.NewRouter()
	restServer.Register(router)

	server := httptest.NewServer(router)
	defer server.Close()

	t.Run("valid api key", func(t *testing.T) {
		body := `{"channel":"test-channel","event":"test-event","payload":"test-payload"}`

		registry.On("Broadcast", mock.MatchedBy(func(msg broadcaster.Message) bool {
			return msg.Channel == "test-channel" && msg.Event == "test-event" && msg.Payload == "test-payload"
		})).Return().Once()

		req, _ := http.NewRequest("POST", server.URL+"/publish", bytes.NewBuffer([]byte(body)))
		req.Header.Set("Authorization", "Bearer test-api-key")
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		registry.AssertExpectations(t)
	})

	t.Run("invalid api key", func(t *testing.T) {
		body := `{"channel":"test-channel","event":"test-event","payload":"test-payload"}`

		req, _ := http.NewRequest("POST", server.URL+"/publish", bytes.NewBuffer([]byte(body)))
		req.Header.Set("Authorization", "Bearer invalid-api-key")
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}
