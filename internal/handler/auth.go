package handler

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/juanpmarin/broadcaster/internal/broadcaster"
)

type AuthRequest struct {
	Token string `json:"token"`
}

type AuthResponse struct {
	Success bool `json:"success"`
}

type AuthHandler struct {
	secret    string
	jwtParser *jwt.Parser
}

func NewAuthHandler(secret string) *AuthHandler {
	jwtParser := jwt.NewParser(
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithLeeway(30*time.Second),
		jwt.WithExpirationRequired(),
		jwt.WithIssuedAt(),
		jwt.WithAudience("broadcaster"),
	)

	return &AuthHandler{
		secret,
		jwtParser,
	}
}

func (h *AuthHandler) Handle(ctx context.Context, req AuthRequest) (AuthResponse, error) {
	claims := JWTClaims{}
	_, err := h.jwtParser.ParseWithClaims(req.Token, &claims, h.keyFunc)
	if err != nil {
		return AuthResponse{}, NewError(ErrorCodeInvalidArgument, err)
	}

	connection, ok := broadcaster.ConnectionFromContext(ctx)
	if !ok {
		return AuthResponse{}, errors.New("connection not found in context")
	}

	if connection.GetUserId() != "" {
		return AuthResponse{}, NewError(ErrorCodeFailedPrecondition, errors.New("connection is already authenticated"))
	}

	userId, err := claims.GetSubject()
	if err != nil || userId == "" {
		return AuthResponse{},
			NewError(ErrorCodeInvalidArgument, errors.New("invalid subject claim"))
	}

	if len(claims.AuthorizedChannels) == 0 {
		return AuthResponse{}, NewError(ErrorCodeInvalidArgument, errors.New("authorized channels cannot be empty"))
	}

	authentication := broadcaster.Authentication{
		UserId:                userId,
		AuthorizedChannelsIds: claims.AuthorizedChannels,
	}

	connection.SetAuthentication(authentication)

	return AuthResponse{
		Success: true,
	}, nil
}

func (h *AuthHandler) keyFunc(token *jwt.Token) (any, error) {
	return []byte(h.secret), nil
}

type JWTClaims struct {
	jwt.RegisteredClaims
	AuthorizedChannels []string `json:"authorizedChannels"`
}

func (c JWTClaims) GetExpirationTime() (*jwt.NumericDate, error) {
	return c.ExpiresAt, nil
}

func (c JWTClaims) GetIssuedAt() (*jwt.NumericDate, error) {
	return c.IssuedAt, nil
}

func (c JWTClaims) GetNotBefore() (*jwt.NumericDate, error) {
	return c.NotBefore, nil
}

func (c JWTClaims) GetIssuer() (string, error) {
	return c.Issuer, nil
}

func (c JWTClaims) GetSubject() (string, error) {
	return c.Subject, nil
}

func (c JWTClaims) GetAudience() (jwt.ClaimStrings, error) {
	return c.Audience, nil
}
