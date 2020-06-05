package provider

import "time"

type Provider interface {
	Get(path string) (Entry, error)
}

type Entry interface {
	Id() string
	Name() string
	Path() string
	Attrs() map[string]interface{}
	HasChildren() bool
	Children() ([]Entry, error)

	AddChild()
	Move()
	Delete()

	Time() time.Time
}
