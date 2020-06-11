package runner

import (
	"context"
	"sync"

	"github.com/grazor/pkdb/pkg/kdb"
)

func Serve(ctx context.Context, providerURI string, serverURIs ...string) (*sync.WaitGroup, error) {
	provider, err := NewProvider(providerURI)
	if err != nil {
		msg := "Unexpected error occurred while creating provider"
		if _, ok := err.(RunnerError); ok {
			msg = err.Error()
		}
		handleError(err, msg)
		return nil, err
	}

	servers, errors := NewServersGroup(serverURIs...)
	if len(errors) > 0 {
		for _, err = range errors {
			msg := "Unexpected error occurred while creating server"
			if _, ok := err.(RunnerError); ok {
				msg = err.Error()
			}
			handleError(err, msg)
		}
		return nil, err
	}

	kdbTree := kdb.New(provider)
	serverErrors, wg, err := servers.Serve(ctx, kdbTree)
	if err != nil {
		msg := "Unexpected error occurred while starting server"
		if _, ok := err.(RunnerError); ok {
			msg = err.Error()
		}
		handleError(err, msg)
		return nil, err
	}

	go func(errors <-chan error) {
		for err := range errors {
			msg := "Unexpected error occurred during server runtime"
			if _, ok := err.(RunnerError); ok {
				msg = err.Error()
			}
			handleError(err, msg)
		}
	}(serverErrors)
	return wg, nil
}
