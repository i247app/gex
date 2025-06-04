package jwtutil

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
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

type Toolkit struct {
	Private *ecdsa.PrivateKey
	Public  *ecdsa.PublicKey
}

func NewJwtToolkit(privateRaw, publicRaw []byte) (*Toolkit, error) {
	priv, err := buildECDSAPrivateKey(privateRaw)
	if err != nil {
		return nil, fmt.Errorf("failed to load ECDSA key: %w", err)
	}

	pub, err := buildECDSAPublicKey(publicRaw)
	if err != nil {
		return nil, fmt.Errorf("failed to load ECDSA key: %w", err)
	}

	return &Toolkit{Private: priv, Public: pub}, nil
}

func (t *Toolkit) SignToken(jwtToken *jwt.Token) (string, error) {
	return jwtToken.SignedString(t.Private)
}

func (t *Toolkit) GenerateJwt(claims jwt.Claims) (*jwt.Token, error) {
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

func (t *Toolkit) StringToToken(tokenString string, claims jwt.Claims) (*jwt.Token, error) {
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return t.Public, nil
	})
	return token, err
}

func (t *Toolkit) GetAuthorizationHeaderJwt(authorizationHeader string) (*jwt.Token, error) {
	if !strings.HasPrefix(authorizationHeader, "Bearer ") {
		return nil, fmt.Errorf("malformed Authorization header, doesn't start with Bearer")
	}
	tokenString := strings.TrimPrefix(authorizationHeader, "Bearer ")
	tok, err := t.StringToToken(tokenString, &CustomClaims{})
	if err != nil {
		return nil, fmt.Errorf("failed to parse JWT token: %w", err)
	}
	return tok, nil
}

func (t *Toolkit) GetRequestJwt(r *http.Request) (*jwt.Token, error) {
	authToken := r.Header.Get("Authorization")
	return t.GetAuthorizationHeaderJwt(authToken)
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
