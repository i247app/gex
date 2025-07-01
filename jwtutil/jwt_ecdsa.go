package jwtutil

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"github.com/golang-jwt/jwt"
)

type EcdsaJwtHelper struct {
	Private *ecdsa.PrivateKey
	Public  *ecdsa.PublicKey
}

func NewEcdsaJwtHelper(privateRaw, publicRaw []byte) (*EcdsaJwtHelper, error) {
	priv, err := buildECDSAPrivateKey(privateRaw)
	if err != nil {
		return nil, fmt.Errorf("failed to load ECDSA key: %w", err)
	}

	pub, err := buildECDSAPublicKey(publicRaw)
	if err != nil {
		return nil, fmt.Errorf("failed to load ECDSA key: %w", err)
	}

	return &EcdsaJwtHelper{Private: priv, Public: pub}, nil
}

func (t *EcdsaJwtHelper) SignToken(jwtToken *jwt.Token) (string, error) {
	return jwtToken.SignedString(t.Private)
}

func (t *EcdsaJwtHelper) GenerateJwt(claims jwt.Claims) (*jwt.Token, error) {
	// Create a new JWT token with ES256 signing method
	jwtToken := jwt.NewWithClaims(
		jwt.SigningMethodES256,
		claims,
	)

	// You can add claims to the token here if needed
	// For example, t.Claims = jwt.MapClaims{"user": "example"}

	// Sign the token using the ECDSA private key
	_, err := jwtToken.SignedString(t.Private)
	if err != nil {
		return nil, fmt.Errorf("failed to sign token: %w", err)
	}

	return jwtToken, nil
}

func (t *EcdsaJwtHelper) StringToToken(tokenString string, claims jwt.Claims) (*jwt.Token, error) {
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return t.Public, nil
	})
	return token, err
}

// buildECDSAPrivateKey loads an ECDSA private key from a PEM file
func buildECDSAPrivateKey(body []byte) (*ecdsa.PrivateKey, error) {
	block, rest := pem.Decode(body)
	if block.Type != "EC PRIVATE KEY" {
		block, _ = pem.Decode(rest)
	}
	if block == nil || block.Type != "EC PRIVATE KEY" {
		return nil, fmt.Errorf("failed to decode PEM block containing the key")
	}

	key, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse EC private key: %w", err)
	}

	return key, nil
}

// buildECDSAPublicKey loads an ECDSA public key from a PEM file
func buildECDSAPublicKey(body []byte) (*ecdsa.PublicKey, error) {
	// Decode the PEM block
	block, _ := pem.Decode(body)
	if block == nil || block.Type != "PUBLIC KEY" {
		return nil, fmt.Errorf("failed to decode PEM block containing the public key")
	}

	// Parse the ECDSA public key
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ECDSA public key: %w", err)
	}

	// Assert that the public key is of type *ecdsa.PublicKey
	ecdsaPub, ok := pub.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not an ECDSA public key")
	}

	return ecdsaPub, nil
}
