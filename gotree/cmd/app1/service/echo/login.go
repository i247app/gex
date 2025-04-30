// cmd/app1/service/login/login.go
package login

import (
	"context"
	"fmt"
	"log"

	"github.com/your-org/gosvr_svr/internal/session" // Correct import path
	"github.com/your-org/gosvr_svr/internal/service" // Correct import path
)

// Request defines the input for the Login method.
type Request struct {
	Username string
	Password string // In a real app, handle this VERY carefully.  Use bcrypt.
}

// Response defines the output for the Login method.
type Response struct {
	AuthToken string
	UserID    string
	Message   string //Added message
}

// Service implements the login service.
type Service struct {
	*service.BaseService // Embed the base service.
	//  Add any login-specific dependencies here (e.g., a user database).
}

// NewService creates a new Login service.
func NewService(baseService *service.BaseService) *Service {
	return &Service{
		BaseService: baseService,
		// Initialize any login-specific dependencies.
	}
}

// Login handles user login.
func (s *Service) Login(ctx context.Context, req *Request) (*Response, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	// 1. Input Validation
	if req.Username == "" || req.Password == "" {
		return nil, service.ErrInvalidRequest
	}

	// 2. Authentication (Replace with secure authentication logic)
	//    -  DO NOT store passwords in plain text.  Use bcrypt or a similar
	//       hashing algorithm.
	//    -  This is a *placeholder* for authentication.
	userID, err := s.authenticateUser(req.Username, req.Password)
	if err != nil {
		return nil, err // Return the error from authentication.
	}

	// 3. Session Management
	session := s.BaseService.SessionManager.CreateSession(userID)

	log.Printf("User %s logged in, session ID: %s", req.Username, session.ID)

	resp := &Response{
		AuthToken: session.ID,
		UserID:    userID,
		Message:   "Login successful",
	}
	return resp, nil
}

// authenticateUser is a *placeholder* for real authentication.
//  DO NOT USE THIS IN PRODUCTION.  It stores passwords in plain text.
func (s *Service) authenticateUser(username, password string) (string, error) {
	// Replace this with a database lookup and bcrypt comparison.
	//  Example (using a hypothetical user database):
	// user, err := s.userRepository.GetUserByUsername(username)
	// if err != nil {
	//   return "", err // Or wrap with fmt.Errorf("login failed: %w", err)
	// }
	// if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
	//   return "", service.ErrUnauthorized // Or a more specific error
	// }
	// return user.ID, nil

	if username == "testuser" && password == "password" {
		return "user123", nil // Hardcoded user for demonstration purposes ONLY.
	}
	return "", service.ErrUnauthorized // Or a more specific error
}
