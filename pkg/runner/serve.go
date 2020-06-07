package runner

import (
	"context"

	"github.com/grazor/pkdb/pkg/kdb"
)

func Serve(ctx context.Context, providerURI string, serverURIs ...string) {
	provider, err := NewProvider(providerURI)
	if err != nil {
		msg := "Unexpected error occurred while creating provider"
		if _, ok := err.(RunnerError); ok {
			msg = err.Error()
		}
		handleError(err, msg)
		return
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
		return
	}

	kdbTree := kdb.New(provider)
	serveCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	serverErrors, wg, err := servers.Serve(serveCtx, kdbTree)
	if err != nil {
		msg := "Unexpected error occurred while starting server"
		if _, ok := err.(RunnerError); ok {
			msg = err.Error()
		}
		handleError(err, msg)
		return
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

	wg.Wait()
}
