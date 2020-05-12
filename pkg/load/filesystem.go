package load

import (
	"os"
	"path/filepath"
	"sync"
)

type FsNode struct {
	*NodeData

	path     string
	parent   *FsNode
	children []*FsNode
}

func NewFsNode(path string) (*FsNode, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return nil, err
	}

	return &FsNode{path: absPath}, nil
}

func (node *FsNode) scanDir(depth int, wg *sync.WaitGroup) error {
	if depth == 0 {
		return nil
	}

	dir, err := os.Open(node.path)
	if err != nil {
		return err
	}

	dirContents, err := dir.Readdir(-1)
	if err != nil {
		return err
	}

	for _, content := range dirContents {
		fullPath := filepath.Join(node.path, content.Name())
		childNode := &FsNode{NodeData: &NodeData{content.Name()}, path: fullPath, parent: node}
		node.children = append(node.children, node)
		if content.IsDir() {
			wg.Add(1)
			go func() {
				defer wg.Done()
				childNode.scanDir(depth-1, wg)
			}()
		}
	}
	return nil
}

func (node *FsNode) Load(depth int) error {
	if depth == 0 {
		return nil
	}

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		node.scanDir(depth-1, &wg)
	}()

	wg.Wait()
	return nil
}

func (node *FsNode) MetaData() *NodeData {
	return node.NodeData
}

func (node *FsNode) Parent() Node {
	return node.parent
}

func (node *FsNode) Children() []Node {
	var children []Node = make([]Node, len(node.children))
	for i, n := range node.children {
		children[i] = n
	}
	return children
}
