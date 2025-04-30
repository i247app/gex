// cmd/app1/service/echo/echo.go
package echo

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/your-org/gosvr_svr/internal/session" // Correct import path
	"github.com/your-org/gosvr_svr/internal/service" // Correct import path
)

// Request defines the input for the Echo method.
type Request struct {
	Message string
}

// Response defines the output for the Echo method.
type Response struct {
	EchoedMessage string
	Timestamp     string
}

// Service implements the echo service.
type Service struct {
	*service.BaseService // Embed the base service to get common functionality.
}

// NewService creates a new Echo service.
func NewService(baseService *service.BaseService) *Service {
	return &Service{
		BaseService: baseService,
	}
}

// Echo echoes the input message and returns a timestamp.
func (s *Service) Echo(ctx context.Context, req *Request) (*Response, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	if req.Message == "" {
		return nil, service.ErrInvalidRequest
	}

	timestamp := time.Now().Format(time.RFC3339)
	echoedMessage := fmt.Sprintf("You said: %s", req.Message)

	log.Printf("Echo: message=%s, timestamp=%s", req.Message, timestamp) // Basic logging.

	resp := &Response{
		EchoedMessage: echoedMessage,
		Timestamp:     timestamp,
	}
	return resp, nil
}
