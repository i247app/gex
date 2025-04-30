// cmd/app1/service/time/time.go
package time

import (
	"context"
	"time"

	"github.com/your-org/gosvr_svr/internal/service" // Correct import path
)

// Request defines the input for the Time method.  It's often empty.
type Request struct{}

// Response defines the output for the Time method.
type Response struct {
	CurrentTime string
}

// Service implements the time service.
type Service struct {
	*service.BaseService // Embed the base service.
}

// NewService creates a new Time service.
func NewService(baseService *service.BaseService) *Service {
	return &Service{
		BaseService: baseService,
	}
}

// Time returns the current time.
func (s *Service) Time(ctx context.Context, req *Request) (*Response, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	now := time.Now().Format(time.RFC3339)
	return &Response{CurrentTime: now}, nil
}
