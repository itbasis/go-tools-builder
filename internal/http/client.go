package http

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

const (
	DefaultDownloadTimeout = 1 * time.Minute

	maxRedirect = 10
)

func NewHTTPClient() *http.Client {
	return &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			slog.Debug(fmt.Sprintf("'%s' redirect to '%s'...", via[0].URL, req.URL))

			if len(via) >= maxRedirect {
				return errors.New("too many redirects")
			}

			return nil
		},
	}
}
