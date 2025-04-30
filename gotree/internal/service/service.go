package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/your-org/gosvr_svr/internal/session"
)

// Service defines the interface for a service.
//
// This interface should define the *business logic* of your service.  It
// should *not* include details about how the service is implemented
// (e.g., database access, caching).  Those details belong in the
// *implementation* of the interface (i.e., in a struct).
//
// Key principles:
//   - Context:  All methods should take a context.Context.
//   - Input/Output:  Use structs for parameters and return values.
//   - Error Handling:  Return errors.  Use wrapped errors for context.
type Service interface {
	// ExampleMethod demonstrates a typical service method.
	//
	// It takes a context and a request struct, and returns a response struct
	// and an error.
	ExampleMethod(ctx context.Context, req *ExampleRequest) (*ExampleResponse, error)

	// Add other service methods here.  For example:
	// UserLogin(ctx context.Context, req *UserLoginRequest) (*UserLoginResponse, error)
	// GetUser(ctx context.Context, req *GetUserRequest) (*GetUserResponse, error)
	// CreateOrder(ctx context.Context, req *CreateOrderRequest) (*CreateOrderResponse, error)
}

// BaseService provides common functionality for services.
//
// This is a *struct*, intended to be embedded in specific service
// implementations.  It provides access to common dependencies.
// It does *NOT* implement the Service interface itself.
type BaseService struct {
	SessionManager *session.SessionManager // Access to session management.
	// Add other common dependencies here, such as:
	//   - Logger
	//   - Database connection pool (if all services use the same one)
	//   - Configuration
}

// NewBaseService creates a new BaseService.  This constructor function
// is important for setting up the dependencies of your service.
func NewBaseService(sessionManager *session.SessionManager) *BaseService {
	return &BaseService{
		SessionManager: sessionManager,
		// Initialize other dependencies here.
	}
}

// ExampleRequest defines the input for ExampleMethod.
type ExampleRequest struct {
	UserID    string
	Data      string
	AuthToken string // Example of including auth token in request
}

// ExampleResponse defines the output for ExampleMethod.
type ExampleResponse struct {
	Result   string
	Status   string
	MetaData map[string]interface{} // Example of returning metadata
}

// =============================================================================
// Example Service Implementation
// =============================================================================

// ExampleService implements the Service interface.
//
// This struct holds the dependencies and state needed by the ExampleService.
// It embeds BaseService to get access to common functionality.
type ExampleService struct {
	*BaseService // Embed BaseService
	// Add dependencies specific to ExampleService here, e.g.:
	// - SomeRepository
	// - SomeOtherClient
}

// NewExampleService creates a new ExampleService.
func NewExampleService(baseService *BaseService) *ExampleService {
	return &ExampleService{
		BaseService: baseService,
		// Initialize ExampleService-specific dependencies.
	}
}

// ExampleMethod implements the Service interface's ExampleMethod.
func (s *ExampleService) ExampleMethod(ctx context.Context, req *ExampleRequest) (*ExampleResponse, error) {
	// 1.  Handle Context
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	// 2.  Authentication/Authorization (if applicable)
	if req.AuthToken != "" {
		sess := s.BaseService.SessionManager.GetSession(req.AuthToken)
		if sess == nil {
			return nil, ErrUnauthorized
		}
		if sess.UserID != req.UserID {
			return nil, ErrForbidden
		}
		log.Printf("User %s is authorized.", req.UserID)
	}

	// 3.  Input Validation
	if req.Data == "" {
		return nil, ErrInvalidRequest
	}

	// 4.  Business Logic
	result := "Processed: " + req.Data

	// 5.  Output
	resp := &ExampleResponse{
		Result: result,
		Status: "success",
		MetaData: map[string]interface{}{
			"processed_at": time.Now().Format(time.RFC3339),
			"user_id":      req.UserID,
		},
	}

	// 6. Logging
	log.Printf("ExampleMethod: data=%s, result=%s", req.Data, result)

	return resp, nil
}

// Define common errors.
var (
	ErrUnauthorized   = fmt.Errorf("unauthorized")
	ErrForbidden      = fmt.Errorf("forbidden")
	ErrInvalidRequest = fmt.Errorf("invalid request")
	ErrNotFound       = fmt.Errorf("not found")
	// Add other common errors here.
)
