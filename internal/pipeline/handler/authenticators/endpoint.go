package authenticators

import (
	"context"
	"io"
	"net/http"
)

type Endpoint interface {
	CreateClient() *http.Client
	CreateRequest(ctx context.Context, body io.Reader) (*http.Request, error)
	SendRequest(ctx context.Context, body io.Reader) ([]byte, error)
}
