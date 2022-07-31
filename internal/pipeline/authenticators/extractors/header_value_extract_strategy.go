package extractors

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/dadrus/heimdall/internal/heimdall"
	"github.com/dadrus/heimdall/internal/x/errorchain"
)

type HeaderValueExtractStrategy struct {
	Name   string
	Prefix string
}

func (es HeaderValueExtractStrategy) GetAuthData(s heimdall.Context) (AuthData, error) {
	if val := s.RequestHeader(es.Name); len(val) != 0 {
		if len(es.Prefix) != 0 && !strings.HasPrefix(val, fmt.Sprintf("%s ", es.Prefix)) {
			return nil, errorchain.NewWithMessagef(heimdall.ErrArgument,
				"'%s' header present, but without required '%s' schema", es.Name, es.Prefix)
		}

		return &headerAuthData{
			name:     es.Name,
			rawValue: val,
			value:    strings.TrimSpace(strings.TrimPrefix(val, es.Prefix)),
		}, nil
	}

	return nil, errorchain.NewWithMessagef(heimdall.ErrArgument, "no '%s' header present", es.Name)
}

type headerAuthData struct {
	name     string
	rawValue string
	value    string
}

func (c *headerAuthData) ApplyTo(req *http.Request) {
	req.Header.Add(c.name, c.rawValue)
}

func (c *headerAuthData) Value() string {
	return c.value
}