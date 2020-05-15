package kdb

import (
	"io"
)

type DataNode struct {
	Name   string
	Path   string
	IsLeaf bool
}

type TreeNode interface {
	MetaData() *DataNode
	Parent() TreeNode
	Children() []TreeNode
}

type AccessableNode interface {
	io.ReadWriteCloser
}

type LoadableNode interface {
	Load(depth int) error
	Watch() (chan interface{}, error)
}

type LoadableTreeNode interface {
	TreeNode
	LoadableNode
}
