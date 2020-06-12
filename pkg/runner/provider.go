package runner

import (
	"fmt"
	"net/url"

	"github.com/grazor/pkdb/pkg/provider"
	"github.com/grazor/pkdb/pkg/provider/fs"
)

func NewProvider(URI string) (provider.Provider, error) {
	u, err := url.Parse(URI)
	if err != nil {
		return nil, RunnerError{
			Inner:   err,
			Message: fmt.Sprintf("failed to parse provider URI %v", URI),
		}
	}

	switch u.Scheme {
	case "", "file":
		{
			provider, err := fs.New(u.Hostname() + u.Path)
			if err != nil {
				return nil, RunnerError{
					Inner:   err,
					Message: fmt.Sprintf("failed to create fs provider from %v%v", u.Hostname(), u.Path),
				}
			}
			return provider, nil
		}
	}

	return nil, RunnerError{Message: fmt.Sprintf("unsupported provider scheme %v", u.Scheme)}
}

