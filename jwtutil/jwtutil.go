package jwtutil

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
)

type CustomClaims struct {
	jwt.StandardClaims
	SessionKey string `json:"session_key"` // Key used to get user's session data
	// Role        string `json:"role"`
	// IsSecure    bool   `json:"is_secure"`
	// Requires2FA bool   `json:"requires_2fa"`
}

func (c CustomClaims) Valid() error {
	return c.StandardClaims.Valid()
}

func (c CustomClaims) SubjectAsInt() int64 {
	sub, err := strconv.ParseInt(c.Subject, 10, 64)
	if err != nil {
		return -1
	}
	return sub
}

func NewClaims(sessionKey string) *CustomClaims {
	now := time.Now()
	return &CustomClaims{
		StandardClaims: jwt.StandardClaims{
			// Subject:   strconv.Itoa(-1),
			Subject:   sessionKey,
			IssuedAt:  now.Unix(),
			ExpiresAt: now.Add(time.Hour * 24 * 14).Unix(),
		},
		SessionKey: sessionKey,
		// Role:        "_default_role_anonymous",
		// IsSecure:    false,
		// Requires2FA: false,
	}
}

type JwtHelper interface {
	SignToken(jwtToken *jwt.Token) (string, error)
	GenerateJwt(claims jwt.Claims) (*jwt.Token, error)
	StringToToken(tokenString string, claims jwt.Claims) (*jwt.Token, error)
}

func GetAuthorizationHeaderJwt(authorizationHeader string, jwtHelper JwtHelper) (*jwt.Token, error) {
	if !strings.HasPrefix(authorizationHeader, "Bearer ") {
		return nil, fmt.Errorf("malformed Authorization header, doesn't start with Bearer")
	}
	tokenString := strings.TrimPrefix(authorizationHeader, "Bearer ")
	tok, err := jwtHelper.StringToToken(tokenString, &CustomClaims{})
	if err != nil {
		return nil, fmt.Errorf("failed to parse JWT token: %w", err)
	}
	return tok, nil
}

func GetRequestJwt(r *http.Request, jwtHelper JwtHelper) (*jwt.Token, error) {
	authToken := r.Header.Get("Authorization")
	return GetAuthorizationHeaderJwt(authToken, jwtHelper)
}
