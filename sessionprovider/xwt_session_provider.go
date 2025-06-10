package sessionprovider

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/i247app/gex/jwtutil"
	"github.com/i247app/gex/session"
)

type XwtResult struct {
	XwtToken   string
	SessionKey string
}

// XwtSessionProvider implements SessionProvider for XWT-based authentication
type XwtSessionProvider struct {
	sessionContainer *session.Container
	jwtToolkit       *jwtutil.Toolkit
	sessionFactory   SessionFactory
	sessionTTL       time.Duration
}

// NewXwtSessionProvider creates a new XWT session provider
func NewXwtSessionProvider(
	sessionContainer *session.Container,
	jwtToolkit *jwtutil.Toolkit,
	sessionFactory SessionFactory,
	sessionTTL time.Duration,
) *XwtSessionProvider {
	return &XwtSessionProvider{
		sessionContainer: sessionContainer,
		jwtToolkit:       jwtToolkit,
		sessionFactory:   sessionFactory,
		sessionTTL:       sessionTTL,
	}
}

// GetSessionWithMetadata implements SessionProvider interface with additional metadata
func (x *XwtSessionProvider) GetSessionFromRequest(r *http.Request) (*SessionResult, error) {
	var didAutoRefresh bool

	// 1. Get or create a XWT token
	xwtResult, err := x.getOrCreateXwtToken(r)
	if err != nil {
		return nil, fmt.Errorf("error getting or creating xwt: %w", err)
	}

	authToken := xwtResult.XwtToken
	sessionKey := authToken

	// 2. Get or create session
	sess, ok := x.sessionContainer.Session(sessionKey)
	if sess == nil || !ok {
		sess, err = x.initNewSession(sessionKey, authToken, "gex.xwt_session_provider")
		if err != nil {
			return nil, fmt.Errorf("error initializing new session: %w", err)
		}
	}

	// 3. Check for expired session and refresh if needed
	isExpired, err := x.isSessionExpired(sess)
	if isExpired || err != nil {
		didAutoRefresh = true
		log(">> XwtSessionProvider: session expired, auto-refreshing...")
		sess, err = x.refreshSession(sess)
		if err != nil {
			return nil, fmt.Errorf("error refreshing expired session: %w", err)
		}
	}

	// 4. Update session touched_at
	sess.Put("touched_at", time.Now())

	// 5. Set auth token in request header for downstream handlers
	if r.Header.Get("Authorization") == "" {
		r.Header.Add("Authorization", "Bearer "+authToken)
	}

	return &SessionResult{
		Session:        sess,
		DidAutoRefresh: didAutoRefresh,
		AuthToken:      authToken,
	}, nil
}

func (x *XwtSessionProvider) getOrCreateXwtToken(r *http.Request) (*XwtResult, error) {
	xwtResult, err := x.getValidXwtFromRequest(r)
	if xwtResult != nil && err == nil {
		return xwtResult, nil
	}

	// Failed to get a valid XWT token from the request
	if xwtResult == nil || xwtResult.XwtToken == "" || xwtResult.SessionKey == "" || err != nil {
		log(">> XwtSessionProvider: WARNING ignoring your xwt token - error getting XWT from request:", err)
	} else {
		log(">> XwtSessionProvider: xwt ok")
	}

	// Create a new XWT token with a new session key
	xwtResult, err = x.createNewXwtToken()
	if err != nil {
		return nil, fmt.Errorf("error creating new XWT token: %v", err)
	}

	return xwtResult, nil
}

func (x *XwtSessionProvider) getValidXwtFromRequest(r *http.Request) (*XwtResult, error) {
	// Validate Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, fmt.Errorf("no Authorization header found")
	}

	if authHeader == "Bearer " {
		return nil, fmt.Errorf("no XWT token found in Authorization header")
	}

	// Get XWT Token
	xwtToken := strings.TrimPrefix(authHeader, "Bearer ")
	sessionKey := xwtToken

	return &XwtResult{
		XwtToken:   xwtToken,
		SessionKey: sessionKey,
	}, nil
}

func (x *XwtSessionProvider) createNewXwtToken() (*XwtResult, error) {
	sessionKey := "n/a" // util.GenerateSessionKey()
	claims := jwtutil.NewClaims(sessionKey)
	jwtToken, err := x.jwtToolkit.GenerateJwt(claims)
	if err != nil {
		return nil, fmt.Errorf("error creating new XWT token: %v", err)
	}

	signedToken, err := x.jwtToolkit.SignToken(jwtToken)
	if err != nil {
		return nil, fmt.Errorf("error signing new XWT token: %v", err)
	}

	return &XwtResult{
		XwtToken:   signedToken,
		SessionKey: signedToken,
	}, nil
}

func (x *XwtSessionProvider) initNewSession(sessionKey string, authToken string, source string) (session.SessionStorer, error) {
	sess, _ := x.sessionContainer.InitSession(sessionKey, x.sessionFactory())
	sess.Put("key", sessionKey)
	sess.Put("source", source)
	sess.Put("token", authToken)
	sess.Put("is_secure", false)

	now := time.Now()
	sess.Put("created_at", now)
	sess.Put("expires_at", now.Add(x.sessionTTL))
	sess.Put("touched_at", now)

	return sess, nil
}

func (x *XwtSessionProvider) refreshSession(sess session.SessionStorer) (session.SessionStorer, error) {
	now := time.Now()
	sess.Put("expires_at", now.Add(x.sessionTTL))
	sess.Put("touched_at", now)

	// Increment refresh count
	refreshCountRaw, ok := sess.Get("refresh_count")
	if !ok {
		sess.Put("refresh_count", 1)
	} else {
		refreshCount, ok := refreshCountRaw.(int)
		if !ok {
			sess.Put("refresh_count", 1)
		} else {
			sess.Put("refresh_count", refreshCount+1)
		}
	}

	return sess, nil
}

func (x *XwtSessionProvider) isSessionExpired(sess session.SessionStorer) (bool, error) {
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
