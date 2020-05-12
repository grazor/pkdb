package load

import (
	"fmt"
	"net/url"
)

type NodeData struct {
	Name string
}

type TreeNode interface {
	MetaData() *NodeData
	Parent() TreeNode
	Children() []TreeNode
}

type LoadableNode interface {
	Load(depth int) error
}

type LoadableTreeNode interface {
	TreeNode
	LoadableNode
}

func GetNode(uri string) (LoadableNode, error) {
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
