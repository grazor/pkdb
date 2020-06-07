package server

import (
	"context"
	"fmt"
	"sync"

	"github.com/grazor/pkdb/pkg/kdb"
)

type ServerError struct {
	Inner   error
	Message string
}

func (err ServerError) Error() string {
	return err.Message
}

type Server interface {
	fmt.Stringer
	Errors() <-chan error
	Serve(ctx context.Context, wg *sync.WaitGroup, tree *kdb.KdbTree) error
}
