package load

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type FsNode struct {
	*NodeData

	path     string
	parent   *FsNode
	children []*FsNode
}

func (node *FsNode) String() string {
	type nodeStruct struct {
		n *FsNode
		d int
	}
	children := []nodeStruct{{node, 0}}
	var info string

	for len(children) > 0 {
		n, d := children[0].n, children[0].d
		info += fmt.Sprintf("\n%v%v (%v)", strings.Repeat("-", d), n.Name, n.path)

		children = children[1:]
		if n.children != nil && len(n.children) > 0 {
			for _, child := range n.children {
				children = append([]nodeStruct{{child, d + 1}}, children...)
			}
		}
	}

	return fmt.Sprintln(info)
}

func NewFsNode(path string) (*FsNode, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return nil, err
	}

	return &FsNode{NodeData: &NodeData{Name: "root"}, path: absPath}, nil
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
		node.children = append(node.children, childNode)
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
