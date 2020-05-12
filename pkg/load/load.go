package load

import (
	"fmt"
	"net/url"
)

type NodeData struct {
	Name string
}

type Node interface {
	MetaData() *NodeData
	Parent() Node
	Children() []Node

	Load(depth int) error
}

func GetNode(uri string) (Node, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	switch u.Scheme {
	case "", "file":
		{
			return NewFsNode(u.Hostname() + u.Path)
		}
	}

	return nil, fmt.Errorf("unexpected scheme %s", u.Scheme)
}
