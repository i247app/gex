// cmd/app1/service/error/error.go
package error

import (
	"errors"
	"fmt"

	"github.com/your-org/gosvr_svr/internal/service" // Import the internal service package
)

// Define service-specific errors.  It's good practice to wrap the
// base errors from the internal/service package to add context.

// ErrUserNotFound indicates that a user was not found.
var ErrUserNotFound = fmt.Errorf("user not found: %w", service.ErrNotFound)

// ErrInvalidCredentials indicates that the provided credentials were invalid.
var ErrInvalidCredentials = fmt.Errorf("invalid credentials: %w", service.ErrUnauthorized)

// ErrSessionExpired indicates that the user's session has expired.
var ErrSessionExpired = fmt.Errorf("session expired: %w", service.ErrUnauthorized)

// ErrFailedToCreateUser indicates that user creation failed.
var ErrFailedToCreateUser = fmt.Errorf("failed to create user: %w", errors.New("failed to create user")) //Wrap a new error

// Add more service-specific errors as needed.  For example:
// var ErrEmailNotVerified = fmt.Errorf("email not verified: %w", service.ErrForbidden)
