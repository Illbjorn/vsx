package gallery

import (
	"net/http"
	"net/url"
)

func New(scheme string, host string) Gallery {
	return Gallery{
		BaseURL: &url.URL{
			Scheme: scheme,
			Host:   host,
		},
	}
}

type Gallery struct {
	// BaseURL holds only the `scheme` and `host` properties of the `url.URL` and
	// is intended for use in constructing runtime request URL strings via the
	// `JoinPath()` method
	BaseURL *url.URL

	// Client is a standard HTTP client with a not-forever timeout applied
	Client *http.Client
}
