package handler

import (
	"context"
	"errors"

	"github.com/juanpmarin/broadcaster/internal/auth"
	"github.com/juanpmarin/broadcaster/internal/broadcaster"
	"github.com/juanpmarin/broadcaster/internal/ierr"
)

type AuthRequest struct {
	Token string `json:"token"`
}

type AuthResponse struct {
	Success bool `json:"success"`
}

type AuthHandlerInterface interface {
	Handle(ctx context.Context, req AuthRequest) (AuthResponse, error)
}

type AuthHandler struct {
	authenticator *auth.Authenticator
}

func NewAuthHandler(authenticator *auth.Authenticator) *AuthHandler {
	return &AuthHandler{
		authenticator,
	}
}

func (h *AuthHandler) Handle(ctx context.Context, req AuthRequest) (AuthResponse, error) {
	authentication, err := h.authenticator.AuthenticateJWT(req.Token)
	if err != nil {
		return AuthResponse{}, err
	}

	if !authentication.IsSubscriber() {
		return AuthResponse{}, ierr.New(ierr.ErrorCodeInvalidArgument, errors.New("invalid token for client authentication"))
	}

	connection, ok := broadcaster.ConnectionFromContext(ctx)
	if !ok {
		return AuthResponse{}, errors.New("connection not found in context")
	}

	if connection.GetUserId() != "" {
		return AuthResponse{}, ierr.New(ierr.ErrorCodeFailedPrecondition, errors.New("connection is already authenticated"))
	}

	connection.SetAuthentication(*authentication)

	return AuthResponse{
		Success: true,
	}, nil
}
