package auth

import (
	"testing"
	"time"

	"github.com/goevery/broadcaster/internal/ierr"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestAuthenticator_AuthenticateJWT(t *testing.T) {
	authenticator := NewAuthenticator("test-secret", []string{"test-api-key"})

	t.Run("valid jwt", func(t *testing.T) {
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

		auth, err := authenticator.AuthenticateJWT(tokenString)

		assert.NoError(t, err)
		assert.NotNil(t, auth)
		assert.Equal(t, "test-user", auth.Subject)
		assert.Equal(t, []string{"test-channel"}, auth.AuthorizedChannelsIds)
		assert.Equal(t, []string{"subscribe"}, auth.Scope)
		assert.False(t, auth.IsAdmin)
	})

	t.Run("invalid jwt signature", func(t *testing.T) {
		claims := jwt.MapClaims{
			"sub":                "test-user",
			"exp":                time.Now().Add(time.Hour).Unix(),
			"iat":                time.Now().Unix(),
			"aud":                "broadcaster",
			"authorizedChannels": []string{"test-channel"},
			"scope":              []string{"subscribe"},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte("invalid-secret"))
		assert.NoError(t, err)

		auth, err := authenticator.AuthenticateJWT(tokenString)

		assert.Error(t, err)
		assert.Nil(t, auth)
		assert.IsType(t, ierr.Error{}, err)
		assert.Equal(t, ierr.ErrorCodeUnauthenticated, err.(ierr.Error).Code)
	})

	t.Run("expired jwt", func(t *testing.T) {
		claims := jwt.MapClaims{
			"sub":                "test-user",
			"exp":                time.Now().Add(-time.Hour).Unix(),
			"iat":                time.Now().Unix(),
			"aud":                "broadcaster",
			"authorizedChannels": []string{"test-channel"},
			"scope":              []string{"subscribe"},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte("test-secret"))
		assert.NoError(t, err)

		auth, err := authenticator.AuthenticateJWT(tokenString)

		assert.Error(t, err)
		assert.Nil(t, auth)
		assert.IsType(t, ierr.Error{}, err)
		assert.Equal(t, ierr.ErrorCodeUnauthenticated, err.(ierr.Error).Code)
	})

	t.Run("missing subject", func(t *testing.T) {
		claims := jwt.MapClaims{
			"exp":                time.Now().Add(time.Hour).Unix(),
			"iat":                time.Now().Unix(),
			"aud":                "broadcaster",
			"authorizedChannels": []string{"test-channel"},
			"scope":              []string{"subscribe"},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte("test-secret"))
		assert.NoError(t, err)

		auth, err := authenticator.AuthenticateJWT(tokenString)

		assert.Error(t, err)
		assert.Nil(t, auth)
		assert.IsType(t, ierr.Error{}, err)
		assert.Equal(t, ierr.ErrorCodeInvalidArgument, err.(ierr.Error).Code)
	})

	t.Run("missing authorized channels", func(t *testing.T) {
		claims := jwt.MapClaims{
			"sub":   "test-user",
			"exp":   time.Now().Add(time.Hour).Unix(),
			"iat":   time.Now().Unix(),
			"aud":   "broadcaster",
			"scope": []string{"subscribe"},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte("test-secret"))
		assert.NoError(t, err)

		auth, err := authenticator.AuthenticateJWT(tokenString)

		assert.Error(t, err)
		assert.Nil(t, auth)
		assert.IsType(t, ierr.Error{}, err)
		assert.Equal(t, ierr.ErrorCodeInvalidArgument, err.(ierr.Error).Code)
	})
}

func TestAuthenticator_AuthenticateAPIKey(t *testing.T) {
	authenticator := NewAuthenticator("test-secret", []string{"test-api-key"})

	t.Run("valid api key", func(t *testing.T) {
		auth, err := authenticator.AuthenticateAPIKey("test-api-key")

		assert.NoError(t, err)
		assert.NotNil(t, auth)
		assert.Equal(t, "api", auth.Subject)
		assert.Equal(t, []string{"publish"}, auth.Scope)
		assert.True(t, auth.IsAdmin)
	})

	t.Run("invalid api key", func(t *testing.T) {
		auth, err := authenticator.AuthenticateAPIKey("invalid-api-key")

		assert.Error(t, err)
		assert.Nil(t, auth)
		assert.IsType(t, ierr.Error{}, err)
		assert.Equal(t, ierr.ErrorCodeUnauthenticated, err.(ierr.Error).Code)
	})
}
