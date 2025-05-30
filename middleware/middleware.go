package middleware

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt"
	"github.com/i247app/gex/jwtutil"
	"github.com/i247app/gex/session"
	"github.com/i247app/gex/util"
)

type RequestToSession func(r *http.Request, sessionService *session.Container, jwtToolkit *jwtutil.Toolkit) (session.SessionStorer, error)

type SessionFactory func() session.SessionStorer

// JwtMiddleware is a middleware that handles JWT authentication and session management.
// It checks for an existing JWT token in the Authorization header, generates a new one if none is found,
// and creates a new session if one doesn't exist.
// It also wraps the response writer to capture the response body.
func JwtMiddleware(
	sessionContainer *session.Container,
	jwtToolkit *jwtutil.Toolkit,
	requestToSession RequestToSession,
	sessionFactory SessionFactory,
) func(http.Handler) http.Handler {
	var log = fmt.Println

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var (
				claims    *jwtutil.CustomClaims
				jwtToken  *jwt.Token
				authToken string
			)

			// Skip entire token and session handling if this header is set
			if r.Header.Get("X-Skip-Session") == "true" {
				next.ServeHTTP(w, r)
				return
			}

			// Check for existing JWT token
			if authHeader := r.Header.Get("Authorization"); authHeader != "" {
				// log(">> JwtMiddleware: found JWT token in Authorization header")

				z, err := jwtToolkit.GetAuthorizationHeaderJwt(authHeader)
				if err != nil {
					log(">> JwtMiddleware: error converting Authorization header to JWT:", err)
				} else {
					jwtToken = z
					claims = jwtToken.Claims.(*jwtutil.CustomClaims)
					authToken = strings.TrimPrefix(authHeader, "Bearer ")
				}
			}

			// If no JWT token found, create a new anonymous session
			if jwtToken == nil {
				// log(">> JwtMiddleware: no JWT token found")

				claims = jwtutil.NewClaims(util.GenerateSessionKey())

				zjwtToken, err := jwtToolkit.GenerateJwt(claims)
				if err != nil {
					log(">> JwtMiddleware: error generating JWT:", err)
				}
				jwtToken = zjwtToken
			}

			// Sign the JWT token
			if authToken == "" {
				var z string
				z, err := jwtToolkit.SignToken(jwtToken)
				if err != nil {
					log(">> JwtMiddleware: error signing JWT:", err)
				}
				authToken = z
			}

			// Check if the session exists
			sess, err := requestToSession(r, sessionContainer, jwtToolkit)
			if sess == nil || err != nil {
				log(">> JwtMiddleware: no session found (most likely no client JWT detected)")

				sess, _ := sessionContainer.InitSession(claims.SessionKey, sessionFactory())
				sess.Put("source", "gex.jwt_middleware")
				sess.Put("token", authToken)
				sess.Put("is_secure", false)

			}

			rwrap := &responseWriterWrapper{
				ResponseWriter: w,
				body:           bytes.NewBuffer(nil),
			}

			// TODO hacky but for now we inject an Authorization header if its missing
			if r.Header.Get("Authorization") == "" {
				r.Header.Add("Authorization", "Bearer "+authToken)
			}

			rwrap.Header().Set("X-Auth-Token", authToken)

			next.ServeHTTP(rwrap, r)

			w.Write(rwrap.body.Bytes())
		})
	}
}
