package middleware

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/i247app/gex/session"
)

var (
	ErrMalformedJwt = errors.New("invalid or malformed JWT")

	DefaultSessionTTL = time.Second * 10
)

type SessionFactory func() session.SessionStorer

type JwtResult struct {
	JwtToken   *jwt.Token
	SessionKey string
	AuthToken  string
}

type XwtResult struct {
	XwtToken   string
	SessionKey string
}
