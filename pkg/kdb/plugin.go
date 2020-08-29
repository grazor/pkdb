package kdb

import "context"

type KdbPlugin interface {
	Init(*KdbTree) error
	ID() string
}

type KdbNodeDiscoverer interface {
	NodeDiscovered()
}

type KdbNodeUpdater interface {
	NodeUpdated(context.Context, *KdbNode, []byte)
}

func (tree *KdbTree) RegisterPlugin(name string, p KdbPlugin) {
	tree.Plugins = append(tree.Plugins, p)
}
