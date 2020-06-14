package provider

import (
	"io"
	"time"
)

type ProviderError struct {
	Inner   error
	Message string
}

func (err ProviderError) Error() string {
	return err.Message
}

type Provider interface {
	Get(path string) (Entry, error)
}

type Entry interface {
	ID() string
	Name() string
	Path() string
	Attrs() map[string]interface{}

	HasChildren() bool
	Children() ([]Entry, error)

	// TODO: move write methods to WritableEntry
	Reader(off int64) (io.ReadCloser, error)
	Writer(off int64) (io.WriteCloser, error)

	AddChild(name string, container bool) (Entry, error)
	Move()
	Delete()

	Size() int64
	Time() time.Time
}
