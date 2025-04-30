// cmd/app1/service/whoami/whoami.go
package whoami

import (
	"context"
	"fmt"

	"github.com/your-org/gosvr_svr/internal/session" // Correct import path
	"github.com/your-org/gosvr_svr/internal/service" // Correct import path
)

// Request defines the input for the WhoAmI method.
type Request struct {
	AuthToken string //  The authentication token (session ID) is required.
}

// Response defines the output for the WhoAmI method.
type Response struct {
	UserID   string
	Username string //  Include the username.
}

// Service implements the whoami service.
type Service struct {
	*service.BaseService // Embed the base service.
	//  In a real application, you might have a user repository here.
	//  userRepository UserRepository
}

// NewService creates a new WhoAmI service.
func NewService(baseService *service.BaseService) *Service {
	return &Service{
		BaseService: baseService,
		//  Initialize the user repository.
	}
}

// WhoAmI returns the user ID and username associated with the given session.
func (s *Service) WhoAmI(ctx context.Context, req *Request) (*Response, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	if req.AuthToken == "" {
		return nil, service.ErrInvalidRequest
	}
	// 1. Get the session using the provided AuthToken.
	session := s.BaseService.SessionManager.GetSession(req.AuthToken)
	if session == nil {
		return nil, service.ErrUnauthorized // Or a more specific error like ErrSessionNotFound
	}

	// 2.  Get the user from the database (replace with your actual logic).
	//  For this example, we'll assume the session contains the user ID.
	userID := session.UserID
	username, err := s.getUsernameFromUserID(userID) //  Call the helper.
	if err != nil {
		return nil, err //  Return the error from the helper.
	}

	return &Response{
		UserID:   userID,
		Username: username,
	}, nil
}

// getUsernameFromUserID is a *placeholder* for retrieving the username
// from a database.  You should replace this with your actual database query.
func (s *Service) getUsernameFromUserID(userID string) (string, error) {
	// Replace this with a database query to get the username.
	// Example:
	// user, err := s.userRepository.GetUserByID(userID)
	// if err != nil {
	//   return "", fmt.Errorf("failed to get username: %w", err)
	// }
	// return user.Username, nil

	switch userID {
	case "user123":
		return "testuser", nil // Hardcoded for the example.
	default:
		return "", fmt.Errorf("user not found: %s", userID) // Use a wrapped error.
	}
}
