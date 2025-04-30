// cmd/app1/service/logout/logout.go
package logout

import (
	"context"
	"log"

	"github.com/your-org/gosvr_svr/internal/session" // Correct import path
	"github.com/your-org/gosvr_svr/internal/service" // Correct import path
)

// Request defines the input for the Logout method.  It needs the
// authentication token (session ID) to identify the session to destroy.
type Request struct {
	AuthToken string
}

// Response defines the output for the Logout method.
type Response struct {
	Message string
}

// Service implements the logout service.
type Service struct {
	*service.BaseService // Embed the base service.
}

// NewService creates a new Logout service.
func NewService(baseService *service.BaseService) *Service {
	return &Service{
		BaseService: baseService,
	}
}

// Logout handles user logout.
func (s *Service) Logout(ctx context.Context, req *Request) (*Response, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	if req.AuthToken == "" {
		return nil, service.ErrInvalidRequest
	}

	// 1.  Get the session.
	session := s.BaseService.SessionManager.GetSession(req.AuthToken)
	if session == nil {
		return nil, service.ErrUnauthorized // Or a more specific error like ErrSessionNotFound
	}

	// 2.  Delete the session.
	s.BaseService.SessionManager.DeleteSession(req.AuthToken)

	log.Printf("User logged out, session ID: %s", req.AuthToken)

	return &Response{Message: "Logout successful"}, nil
}
