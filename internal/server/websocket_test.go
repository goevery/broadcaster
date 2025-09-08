package server

import (
	"encoding/json"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/goevery/broadcaster/internal/auth"
	"github.com/goevery/broadcaster/internal/broadcaster"
	"github.com/goevery/broadcaster/internal/handler"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestWebSocketServer(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	registry := broadcaster.NewInMemoryRegistry(logger)
	authenticator := auth.NewAuthenticator("test-secret", []string{"test-api-key"})
	channelIdValidator := handler.NewChannelIdValidator()
	heartbeatHandler := handler.NewHeartbeatHandler()
	joinHandler := handler.NewJoinHandler(channelIdValidator, registry)
	leaveHandler := handler.NewLeaveHandler(channelIdValidator, registry)
	pushHandler := handler.NewPushHandler(channelIdValidator, registry)
	authHandler := handler.NewAuthHandler(authenticator)

	router := NewRouter(logger, heartbeatHandler, joinHandler, leaveHandler, pushHandler, authHandler)
	upgrader := &websocket.Upgrader{}

	wsServer := NewWebSocketServer(logger, upgrader, registry, router)

	mainRouter := mux.NewRouter()
	wsServer.Register(mainRouter)

	server := httptest.NewServer(mainRouter)
	defer server.Close()

	u, _ := url.Parse(server.URL)
	u.Scheme = "ws"
	u.Path = "/websocket"

	t.Run("successful flow", func(t *testing.T) {
		conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
		assert.NoError(t, err)

		// Auth
		claims := jwt.MapClaims{
			"sub":                "test-user",
			"exp":                time.Now().Add(time.Hour).Unix(),
			"iat":                time.Now().Unix(),
			"aud":                "broadcaster",
			"authorizedChannels": []string{"test-channel"},
			"scope":              []string{"subscribe"},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte("test-secret"))
		assert.NoError(t, err)

		authRequest := json.RawMessage(`{"id":1,"method":"auth","params":{"token":"` + tokenString + `"}}`)
		err = conn.WriteJSON(authRequest)
		assert.NoError(t, err)

		var authResponse handler.Response
		conn.SetReadDeadline(time.Now().Add(time.Second))
		err = conn.ReadJSON(&authResponse)
		assert.NoError(t, err)

		var authResponsePayload handler.AuthResponse
		err = json.Unmarshal(*authResponse.Result, &authResponsePayload)
		assert.NoError(t, err)
		assert.Equal(t, true, authResponsePayload.Success)

		// Join
		joinRequest := json.RawMessage(`{"id":2,"method":"join","params":{"channelId":"test-channel"}}`)
		err = conn.WriteJSON(joinRequest)
		assert.NoError(t, err)

		var joinResponse handler.Response
		conn.SetReadDeadline(time.Now().Add(time.Second))
		err = conn.ReadJSON(&joinResponse)
		assert.NoError(t, err)

		var joinResponsePayload handler.JoinResponse
		err = json.Unmarshal(*joinResponse.Result, &joinResponsePayload)
		assert.NoError(t, err)
		assert.NotEmpty(t, joinResponsePayload.SubscriptionId)

		// Server sends a message
		msg := broadcaster.Message{ChannelId: "test-channel", Payload: "test-payload"}
		registry.Broadcast(msg)

		var messageRequest handler.Request
		conn.SetReadDeadline(time.Now().Add(time.Second))
		err = conn.ReadJSON(&messageRequest)
		assert.NoError(t, err)
		assert.Equal(t, "broadcast", messageRequest.Method)

		var messagePayload broadcaster.Message
		err = json.Unmarshal(*messageRequest.Params, &messagePayload)
		assert.NoError(t, err)
		assert.Equal(t, msg.ChannelId, messagePayload.ChannelId)
		assert.Equal(t, msg.Payload, messagePayload.Payload)

		conn.Close()
	})

	t.Run("invalid message", func(t *testing.T) {
		conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
		assert.NoError(t, err)
		defer conn.Close()

		err = conn.WriteMessage(websocket.TextMessage, []byte("invalid-json"))
		assert.NoError(t, err)

		// The server should close the connection
		conn.SetReadDeadline(time.Now().Add(time.Second * 10))
		_, _, err = conn.ReadMessage()
		assert.Error(t, err)
		assert.True(t, websocket.IsCloseError(err, websocket.CloseNoStatusReceived))
	})

	t.Run("join without auth", func(t *testing.T) {
		conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
		assert.NoError(t, err)
		defer conn.Close()

		joinRequest := json.RawMessage(`{"id":1,"method":"join","params":{"channelId":"test-channel"}}`)
		err = conn.WriteJSON(joinRequest)
		assert.NoError(t, err)

		var joinResponse handler.Response
		conn.SetReadDeadline(time.Now().Add(time.Second))
		err = conn.ReadJSON(&joinResponse)
		assert.NoError(t, err)
		assert.NotNil(t, joinResponse.Error)
		assert.Equal(t, "Unauthenticated", string(joinResponse.Error.Code))
	})

	t.Run("join unauthorized channel", func(t *testing.T) {
		conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
		assert.NoError(t, err)
		defer conn.Close()

		// Auth
		claims := jwt.MapClaims{
			"sub":                "test-user",
			"exp":                time.Now().Add(time.Hour).Unix(),
			"iat":                time.Now().Unix(),
			"aud":                "broadcaster",
			"authorizedChannels": []string{"another-channel"},
			"scope":              []string{"subscribe"},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte("test-secret"))
		assert.NoError(t, err)

		authRequest := json.RawMessage(`{"id":1,"method":"auth","params":{"token":"` + tokenString + `"}}`)
		err = conn.WriteJSON(authRequest)
		assert.NoError(t, err)

		var authResponse handler.Response
		conn.SetReadDeadline(time.Now().Add(time.Second))
		err = conn.ReadJSON(&authResponse)
		assert.NoError(t, err)

		// Join
		joinRequest := json.RawMessage(`{"id":2,"method":"join","params":{"channelId":"test-channel"}}`)
		err = conn.WriteJSON(joinRequest)
		assert.NoError(t, err)

		var joinResponse handler.Response
		conn.SetReadDeadline(time.Now().Add(time.Second))
		err = conn.ReadJSON(&joinResponse)
		assert.NoError(t, err)
		assert.NotNil(t, joinResponse.Error)
		assert.Equal(t, "Unauthenticated", string(joinResponse.Error.Code))
	})

	t.Run("push message with publish scope", func(t *testing.T) {
		conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
		assert.NoError(t, err)
		defer conn.Close()

		// Auth
		claims := jwt.MapClaims{
			"sub":                "test-user",
			"exp":                time.Now().Add(time.Hour).Unix(),
			"iat":                time.Now().Unix(),
			"aud":                "broadcaster",
			"authorizedChannels": []string{"test-channel"},
			"scope":              []string{"publish"},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte("test-secret"))
		assert.NoError(t, err)

		authRequest := json.RawMessage(`{"id":1,"method":"auth","params":{"token":"` + tokenString + `"}}`)
		err = conn.WriteJSON(authRequest)
		assert.NoError(t, err)

		var authResponse handler.Response
		conn.SetReadDeadline(time.Now().Add(time.Second))
		err = conn.ReadJSON(&authResponse)
		assert.NoError(t, err)

		var authResponsePayload handler.AuthResponse
		err = json.Unmarshal(*authResponse.Result, &authResponsePayload)
		assert.NoError(t, err)
		assert.Equal(t, true, authResponsePayload.Success)

		// Push
		pushRequest := json.RawMessage(`{"id":2,"method":"push","params":{"channelId":"test-channel","payload":{"foo":"bar"}}}`)
		err = conn.WriteJSON(pushRequest)
		assert.NoError(t, err)

		var pushResponse handler.Response
		conn.SetReadDeadline(time.Now().Add(time.Second))
		err = conn.ReadJSON(&pushResponse)
		assert.NoError(t, err)
		assert.Nil(t, pushResponse.Error)
	})

	t.Run("join without subscribe scope", func(t *testing.T) {
		conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
		assert.NoError(t, err)
		defer conn.Close()

		// Auth with only publish scope (no subscribe)
		claims := jwt.MapClaims{
			"sub":                "test-user",
			"exp":                time.Now().Add(time.Hour).Unix(),
			"iat":                time.Now().Unix(),
			"aud":                "broadcaster",
			"authorizedChannels": []string{"test-channel"},
			"scope":              []string{"publish"},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte("test-secret"))
		assert.NoError(t, err)

		authRequest := json.RawMessage(`{"id":1,"method":"auth","params":{"token":"` + tokenString + `"}}`)
		err = conn.WriteJSON(authRequest)
		assert.NoError(t, err)

		var authResponse handler.Response
		conn.SetReadDeadline(time.Now().Add(time.Second))
		err = conn.ReadJSON(&authResponse)
		assert.NoError(t, err)

		var authResponsePayload handler.AuthResponse
		err = json.Unmarshal(*authResponse.Result, &authResponsePayload)
		assert.NoError(t, err)
		assert.Equal(t, true, authResponsePayload.Success)

		// Join should fail without subscribe scope
		joinRequest := json.RawMessage(`{"id":2,"method":"join","params":{"channelId":"test-channel"}}`)
		err = conn.WriteJSON(joinRequest)
		assert.NoError(t, err)

		var joinResponse handler.Response
		conn.SetReadDeadline(time.Now().Add(time.Second))
		err = conn.ReadJSON(&joinResponse)
		assert.NoError(t, err)
		assert.NotNil(t, joinResponse.Error)
		assert.Equal(t, "PermissionDenied", string(joinResponse.Error.Code))
	})

	t.Run("push message without publish scope", func(t *testing.T) {
		conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
		assert.NoError(t, err)
		defer conn.Close()

		// Auth with only subscribe scope (no publish)
		claims := jwt.MapClaims{
			"sub":                "test-user",
			"exp":                time.Now().Add(time.Hour).Unix(),
			"iat":                time.Now().Unix(),
			"aud":                "broadcaster",
			"authorizedChannels": []string{"test-channel"},
			"scope":              []string{"subscribe"},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte("test-secret"))
		assert.NoError(t, err)

		authRequest := json.RawMessage(`{"id":1,"method":"auth","params":{"token":"` + tokenString + `"}}`)
		err = conn.WriteJSON(authRequest)
		assert.NoError(t, err)

		var authResponse handler.Response
		conn.SetReadDeadline(time.Now().Add(time.Second))
		err = conn.ReadJSON(&authResponse)
		assert.NoError(t, err)

		var authResponsePayload handler.AuthResponse
		err = json.Unmarshal(*authResponse.Result, &authResponsePayload)
		assert.NoError(t, err)
		assert.Equal(t, true, authResponsePayload.Success)

		// Push should fail without publish scope
		pushRequest := json.RawMessage(`{"id":2,"method":"push","params":{"channelId":"test-channel","payload":{"foo":"bar"}}}`)
		err = conn.WriteJSON(pushRequest)
		assert.NoError(t, err)

		var pushResponse handler.Response
		conn.SetReadDeadline(time.Now().Add(time.Second))
		err = conn.ReadJSON(&pushResponse)

		assert.NoError(t, err)
		assert.NotNil(t, pushResponse.Error)
		assert.Equal(t, "PermissionDenied", string(pushResponse.Error.Code))
	})
}
