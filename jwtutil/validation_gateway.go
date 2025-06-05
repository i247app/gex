package jwtutil

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/i247app/gex/session"
)

const (
	DefaultSessionTTL = time.Second * 10
)

var (
	log             = fmt.Println
	ErrMalformedJwt = errors.New("invalid or malformed JWT")
)

type ValidationGateway struct {
	sessionContainer *session.Container
	Toolkit          *Toolkit
}

func NewValidationGateway(toolkit *Toolkit, sessionContainer *session.Container) *ValidationGateway {
	return &ValidationGateway{
		Toolkit:          toolkit,
		sessionContainer: sessionContainer,
	}
}

type ValidationResult struct {
	Tag              string
	Intent           string
	IsJwtExist       bool
	IsJwtValid       bool
	IsSessionExist   bool
	IsSessionExpired bool
}

func (v *ValidationGateway) Validate(r *http.Request) (ValidationResult, error) {
	var (
		sessionKey       string
		intent           string
		isJwtExist       bool
		isJwtValid       bool
		isSessionExist   bool
		isSessionExpired bool
	)

	// JWT Middleware Flow:
	//
	// 0. Check for Authorization header
	// 1. Check for valid JWT token based on the Authorization header, set result to isJwtExist
	// 2a. If JWT token found, just get the session key
	// 2b. If JWT token NOT found, create a new JWT token with session key
	// 3. Check for a session linked to the JWT token's sessionKey
	// 3a. If session is found, do nothing
	// 3b. 🚨 If session is not found AND isJwtExist == true, return unauthorized error
	// 3c. If session is not found AND isJwtExist == false, create a new session
	// 4. Check session expiration
	// 4a. 🚨 If session is expired, return unauthorized error
	// 4b. If session is not expired, do nothing
	// 5. Update session touched_at
	// 6. Determine current authToken
	// 7. Set the session token in the Authorization header

	// 0. Check for Authorization header
	isJwtExist = r.Header.Get("Authorization") != ""
	if !isJwtExist {
		log(">> ValidationGateway: no JWT token found in Authorization header")
		return ValidationResult{
			Tag:              "no Authorization header found",
			Intent:           "maybe an anon API call",
			IsJwtExist:       false,
			IsJwtValid:       false,
			IsSessionExist:   false,
			IsSessionExpired: false,
		}, nil
	}

	// 1. Check for valid JWT token based on the Authorization header
	var incomingSessionKey *string
	jwtToken, err := getJwtFromRequest(r, v.Toolkit)
	// if err == ErrMalformedJwt {
	// 	log(">> ValidationGateway: totally malformed JWT token, returning unauthorized...")
	// 	return ValidationResult{Tag: "malformed jwt"}, fmt.Errorf("Unauthorized")
	// } else if jwtToken == nil || err != nil {
	// 	log(">> ValidationGateway: error getting JWT from request:", err)
	// } else {
	// 	claims, ok := jwtToken.Claims.(*CustomClaims)
	// 	if ok {
	// 		incomingSessionKey = &claims.SessionKey
	// 	}
	// }
	isJwtValid = err == nil && jwtToken != nil && incomingSessionKey != nil
	if !isJwtValid {
		log(">> ValidationGateway: no valid JWT token found, creating a new one...")

		return ValidationResult{
			Tag:              "no valid JWT token found",
			Intent:           "INIT_NEW_JWT",
			IsJwtExist:       true,
			IsJwtValid:       false,
			IsSessionExist:   false,
			IsSessionExpired: false,
		}, nil
	} else {
		// 2a. If JWT token found, just get the session key

		log(">> ValidationGateway: found valid JWT token, using session key from claims...")

		claims, ok := jwtToken.Claims.(*CustomClaims)
		if ok {
			incomingSessionKey = &claims.SessionKey
		}
		sessionKey = *incomingSessionKey
	}

	// 3. Check for a session linked to the JWT token's sessionKey
	sess, ok := v.sessionContainer.Session(sessionKey)
	isSessionExist = ok && sess != nil

	// 3a. If session is found, do nothing

	// 3b. 🚨 If session is not found AND isJwtExist == true, return unauthorized error
	if !isSessionExist {
		if isJwtValid {

			log(">> ValidationGateway: JWT token exists but no session found, returning unauthorized...")
			return ValidationResult{
				Tag:              "no session found but jwt exists & valid",
				IsJwtExist:       true,
				IsJwtValid:       true,
				IsSessionExist:   false,
				IsSessionExpired: false,
			}, fmt.Errorf("Unauthorized")
		}
		// 3c. If session is not found AND isJwtExist == false, create a new session

		log(">> ValidationGateway: no session found (most likely no client JWT detected)")

		// Initialize a new session
		intent = "INIT_NEW_SESSION"
		// if authToken != "" {
		// 	sess, _ = initNewSession(sessionKey, authToken, v.sessionContainer, sessionFactory, sessionTTL)
		// 	if sess == nil {
		// 		log(">> ValidationGateway: error initializing new session")
		// 	}
		// }
	}

	if sess == nil {
		// !! Major error, just return unauthorized
		return ValidationResult{
			Tag:              "attempted to create new session but failed",
			IsJwtExist:       true,
			IsJwtValid:       true,
			IsSessionExist:   false,
			IsSessionExpired: false,
		}, fmt.Errorf("Unauthorized")
	}

	// 4. Check session expiration
	tmp, err := checkSessionExpired(sess)
	isSessionExpired = tmp
	if isSessionExpired || err != nil {
		// didAutoRefresh = true

		log(">> ValidationGateway: session expired, for now just auto-refreshing...")

		// Initialize a new session
		intent = "REFRESH_EXPIRED_SESSION"
		// sess, _ = refreshSession(sess, sessionTTL)
		// if sess == nil {
		// 	log(">> ValidationGateway: error initializing new session")
		// }
	} else {
		// 4b. If session is not expired, do nothing
	}

	return ValidationResult{
		Tag:              "session validation completed",
		Intent:           intent,
		IsJwtExist:       isJwtExist,
		IsJwtValid:       isJwtValid,
		IsSessionExist:   isSessionExist,
		IsSessionExpired: isSessionExpired,
	}, nil
}

func getJwtFromRequest(r *http.Request, jwtToolkit *Toolkit) (*jwt.Token, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, fmt.Errorf("no Authorization header found")
	}

	if authHeader == "Bearer " {
		return nil, fmt.Errorf("no JWT token found in Authorization header")
	}

	// log(">> ValidationGateway: found JWT token in Authorization header")
	jwtToken, err := jwtToolkit.GetAuthorizationHeaderJwt(authHeader)
	if err != nil {
		log(">> ValidationGateway: error converting Authorization header to JWT:", err)
		return nil, ErrMalformedJwt
	}

	return jwtToken, err
}

func checkSessionExpired(sess session.SessionStorer) (bool, error) {
	expiresAtRaw, ok := sess.Get("expires_at")
	if !ok {
		return false, fmt.Errorf("no expires_at found in session")
	}

	expiresAt, ok := expiresAtRaw.(time.Time)
	if !ok {
		return false, fmt.Errorf("error converting expires_at to time.Time")
	}

	return expiresAt.Before(time.Now()), nil
}
