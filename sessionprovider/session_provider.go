package sessionprovider

import (
	"fmt"
	"net/http"
)

var log = fmt.Println

type SessionProvider interface {
	GetSessionFromRequest(r *http.Request) (*SessionResult, error)
}
