package runner

import (
	"context"
	"fmt"
	"net/url"
	"sync"

	"github.com/grazor/pkdb/pkg/kdb"
	"github.com/grazor/pkdb/pkg/server"
	"github.com/grazor/pkdb/pkg/server/fuse"
)

type serversGroup []server.Server

func NewServer(URI string) (server.Server, error) {
	u, err := url.Parse(URI)
	if err != nil {
		return nil, RunnerError{
			Inner:   err,
			Message: fmt.Sprintf("failed to parse server URL %v", URI),
		}
	}

	switch u.Scheme {
	case "", "file":
		{
			var options map[string][]string = u.Query()
			server, err := fuse.New(u.Hostname()+u.Path, options)
			if err != nil {
				return nil, RunnerError{
					Inner:   err,
					Message: fmt.Sprintf("failed to create fuse server for %v%v", u.Hostname(), u.Path),
				}
			}
			return server, nil
		}
	}

	return nil, RunnerError{Message: fmt.Sprintf("unsupported server scheme %v", u.Scheme)}
}

func NewServersGroup(URIs ...string) (serversGroup, []error) {
	providers := make(serversGroup, 0, len(URIs))
	errors := make([]error, 0)

	for _, URI := range URIs {
		provider, err := NewServer(URI)
		if err != nil {
			errors = append(errors, err)
		} else {
			providers = append(providers, provider)
		}
	}

	return providers, errors
}

func (g serversGroup) Serve(ctx context.Context, tree *kdb.KdbTree) (<-chan error, *sync.WaitGroup, error) {
	processErrors := func(ctx context.Context, serv server.Server, errors chan<- error) {
		for {
			select {
			case <-ctx.Done():
				return
			case err := <-serv.Errors():
				msg := "unknown server message occurred"
				if _, ok := err.(server.ServerError); ok {
					msg = err.Error()
				}

				select {
				case <-ctx.Done():
					return
				case errors <- RunnerError{
					Inner:   err,
					Message: msg,
					Source:  serv.String(),
				}:
				}

			}
		}

	}

	var wg, errorsWg sync.WaitGroup
	var serveError error
	errors := make(chan error)
	for _, s := range g {
		err := s.Serve(ctx, &wg, tree)
		if err != nil {
			serveError = RunnerError{
				Inner:   err,
				Message: "failed to start server",
				Source:  s.String(),
			}
			break
		}
		errorsWg.Add(1)
		go func(s server.Server) {
			defer errorsWg.Done()
			processErrors(ctx, s, errors)
		}(s)
	}

	go func() {
		errorsWg.Wait()
		close(errors)
	}()

	return errors, &wg, serveError
}
