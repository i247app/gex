package jwtutil

import (
	"fmt"

	"github.com/golang-jwt/jwt"
)

type HmacJwtHelper struct {
	Key []byte
}

func NewHmacJwtHelper(key []byte) (*HmacJwtHelper, error) {
	if len(key) == 0 {
		return nil, fmt.Errorf("key is empty")
	}

	return &HmacJwtHelper{Key: key}, nil
}

func (t *HmacJwtHelper) SignToken(jwtToken *jwt.Token) (string, error) {
	return jwtToken.SignedString(t.Key)
}

func (t *HmacJwtHelper) GenerateJwt(claims jwt.Claims) (*jwt.Token, error) {
	// Create a new JWT token with ES256 signing method
	jwtToken := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	)

	// You can add claims to the token here if needed
	// For example, t.Claims = jwt.MapClaims{"user": "example"}

	// Sign the token using the ECDSA private key
	_, err := jwtToken.SignedString(t.Key)
	if err != nil {
		return nil, fmt.Errorf("failed to sign token: %w", err)
	}

	return jwtToken, nil
}

func (t *HmacJwtHelper) StringToToken(tokenString string, claims jwt.Claims) (*jwt.Token, error) {
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return t.Key, nil
	})
	return token, err
}
