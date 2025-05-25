package jwtutil

import (
	"fmt"
	"net/http"
)

// JwtTokenLocator is a concrete implementation of the AuthTokenLocator interface
// It locates the session token in the Authorization header of the request
type JwtTokenLocator struct {
	jwtToolkit *Toolkit
}

func NewJwtTokenLocator(jwtToolkit *Toolkit) *JwtTokenLocator {
	return &JwtTokenLocator{jwtToolkit: jwtToolkit}
}

func (jtl *JwtTokenLocator) Locate(r *http.Request) (string, error) {
	// Get JWT token
	jwtToken, err := jtl.jwtToolkit.GetAuthorizationHeaderJwt(r.Header.Get("Authorization"))
	if err != nil {
		return "", err
	}

	// Get sessionKey
	claims, ok := jwtToken.Claims.(*CustomClaims)
	if !ok {
		return "", fmt.Errorf("jwt claims could not be cast to CustomClaims")
	}

	return claims.SessionKey, nil
}
