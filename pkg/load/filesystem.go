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
		info += fmt.Sprintf("\n%v-%v (%v)", strings.Repeat(" |", d+1), n.Name, n.path)

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

func (node *FsNode) Load(depth int) error {
	if depth == 0 {
		return nil
	}

	var scanDir func(node *FsNode, depth int, wg *sync.WaitGroup)
	scanDir = func(scanNode *FsNode, depth int, wg *sync.WaitGroup) {
		defer wg.Done()
		if depth == 0 {
			return
		}

		dir, err := os.Open(scanNode.path)
		if err != nil {
			return
		}
		defer dir.Close()

		dirContents, err := dir.Readdir(-1)
		if err != nil {
			return
		}

		if len(dirContents) > 0 {
			scanNode.children = make([]*FsNode, 0, len(dirContents))
		}

		for _, content := range dirContents {
			fullPath := filepath.Join(scanNode.path, content.Name())
			childNode := &FsNode{NodeData: &NodeData{content.Name()}, path: fullPath, parent: scanNode}
			scanNode.children = append(scanNode.children, childNode)
			if content.IsDir() {
				wg.Add(1)
				go scanDir(childNode, depth-1, wg)
			}
		}
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go scanDir(node, depth-1, &wg)
	wg.Wait()
	return nil
}

func (node *FsNode) MetaData() *NodeData {
	return node.NodeData
}

func (node *FsNode) Parent() TreeNode {
	return node.parent
}

func (node *FsNode) Children() []TreeNode {
	var children []TreeNode = make([]TreeNode, len(node.children))
	for i, n := range node.children {
		children[i] = n
	}
	return children
}
