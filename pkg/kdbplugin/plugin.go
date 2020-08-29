// Package kdbplugin defines interfaces for pkdb plugins.
package kdbplugin

type KdbPlugin interface {
	Init() error
	ID() string
}

type KdbNodeDiscoverer interface {
	NodeDiscovered()
}
