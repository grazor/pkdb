package filesystem

import (
	"os"
	"path/filepath"

	"github.com/grazor/pkdb/pkg/kdb"
)

type FsNode struct {
	*kdb.DataNode

	path     string
	parent   *FsNode
	children []*FsNode

	index map[string]*FsNode
}

func New(path string) (*FsNode, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return nil, err
	}

	index := make(map[string]*FsNode)
	node := &FsNode{DataNode: &kdb.DataNode{Name: "root", Path: "/"}, path: absPath}
	node.index = index
	return node, nil
}

func (node *FsNode) MetaData() *kdb.DataNode {
	return node.DataNode
}

func (node *FsNode) Parent() kdb.TreeNode {
	return node.parent
}

func (node *FsNode) Children() []kdb.TreeNode {
	var children []kdb.TreeNode = make([]kdb.TreeNode, len(node.children))
	for i, n := range node.children {
		children[i] = n
	}
	return children
}

