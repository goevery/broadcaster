package auth

import (
	"context"
	"crypto/subtle"
	"errors"
	"slices"
	"time"

	"github.com/goevery/broadcaster/internal/ierr"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	jwt.RegisteredClaims
	AuthorizedChannels []string `json:"authorizedChannels,omitempty"`
	Scope              []string `json:"scope,omitempty"`
}

type Authentication struct {
	Subject            string
	AuthorizedChannels []string
	Scope              []string
	IsAdmin            bool
}

func (a *Authentication) IsPublisher() bool {
	return slices.Contains(a.Scope, "publish")
}

func (a *Authentication) IsSubscriber() bool {
	return slices.Contains(a.Scope, "subscribe")
}

func (a *Authentication) IsAuthorized(channel string) bool {
	if a.Subject == "" {
		return false
	}

	if a.IsAdmin {
		return true
	}

	return slices.Contains(a.AuthorizedChannels, channel)
}

type contextKey string

const authenticationKey contextKey = "authentication"

func WithAuthentication(ctx context.Context, auth *Authentication) context.Context {
	return context.WithValue(ctx, authenticationKey, auth)
}

func AuthenticationFromContext(ctx context.Context) (*Authentication, bool) {
	auth, ok := ctx.Value(authenticationKey).(*Authentication)
	return auth, ok
}

type Authenticator struct {
	secret    []byte
	apiKeys   []string
	jwtParser *jwt.Parser
}

func NewAuthenticator(secret string, apiKeys []string) *Authenticator {
	jwtParser := jwt.NewParser(
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithLeeway(30*time.Second),
		jwt.WithExpirationRequired(),
		jwt.WithIssuedAt(),
		jwt.WithAudience("broadcaster"),
	)

	return &Authenticator{
		secret:    []byte(secret),
		apiKeys:   apiKeys,
		jwtParser: jwtParser,
	}
}

func (a *Authenticator) keyFunc(token *jwt.Token) (any, error) {
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, ierr.New(ierr.ErrorCodeUnauthenticated, errors.New("unexpected signing method"))
	}
	return a.secret, nil
}

func (a *Authenticator) AuthenticateJWT(tokenString string) (*Authentication, error) {
	claims := Claims{}

	_, err := a.jwtParser.ParseWithClaims(tokenString, &claims, a.keyFunc)
	if err != nil {
		return nil, ierr.New(ierr.ErrorCodeUnauthenticated, err)
	}

	subject, err := claims.GetSubject()
	if err != nil || subject == "" {
		return nil, ierr.New(ierr.ErrorCodeInvalidArgument, errors.New("invalid subject claim"))
	}

	if len(claims.AuthorizedChannels) == 0 {
		return nil, ierr.New(ierr.ErrorCodeInvalidArgument, errors.New("authorized channels cannot be empty"))
	}

	return &Authentication{
		Subject:            subject,
		AuthorizedChannels: claims.AuthorizedChannels,
		Scope:              claims.Scope,
		IsAdmin:            false,
	}, nil
}

func (a *Authenticator) AuthenticateAPIKey(apiKey string) (*Authentication, error) {
	for _, key := range a.apiKeys {
		if subtle.ConstantTimeCompare([]byte(apiKey), []byte(key)) == 1 {
			return &Authentication{
				Subject: "api",
				Scope:   []string{"publish"},
				IsAdmin: true,
			}, nil
		}
	}

	return nil, ierr.New(ierr.ErrorCodeUnauthenticated, errors.New("invalid api key"))
}
