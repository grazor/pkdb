package load

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
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

	index map[string]*FsNode
}

func (node *FsNode) String() string {
	return fmt.Sprintf("%v: %v\n", node.path, node.Name)
}

func (node *FsNode) GetTree() string {
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

	index := make(map[string]*FsNode)
	node := &FsNode{NodeData: &NodeData{Name: "root"}, path: absPath}
	node.index = index
	return node, nil
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

func (node *FsNode) Load(depth int) error {
	if depth == 0 {
		return nil
	}

	var scanDir func(*FsNode, int, *sync.WaitGroup, chan<- *FsNode)
	scanDir = func(scanNode *FsNode, depth int, wg *sync.WaitGroup, found chan<- *FsNode) {
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

		scanNode.children = make([]*FsNode, 0, len(dirContents))
		for _, content := range dirContents {
			fullPath := filepath.Join(scanNode.path, content.Name())
			childNode := &FsNode{NodeData: &NodeData{content.Name()}, path: fullPath, parent: scanNode, index: scanNode.index}
			found <- childNode
			scanNode.children = append(scanNode.children, childNode)
			if content.IsDir() {
				wg.Add(1)
				go scanDir(childNode, depth-1, wg, found)
			}
		}
	}

	indexNodes := func(index map[string]*FsNode, found <-chan *FsNode, done chan interface{}) {
		for {
			select {
			case node := <-found:
				index[node.path] = node
			case <-done:
				return
			}
		}
	}

	var wg sync.WaitGroup
	found := make(chan *FsNode, 32)
	done := make(chan interface{})

	wg.Add(1)
	go scanDir(node, depth-1, &wg, found)
	go indexNodes(node.index, found, done)

	wg.Wait()
	close(done)

	return nil
}

func (node *FsNode) Watch() (chan interface{}, error) {
	watch := func(watcher *fsnotify.Watcher, done chan interface{}, root *FsNode) {
		defer watcher.Close()
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Chmod != fsnotify.Chmod {
					// TODO: handle deletion
					nodePath := filepath.Dir(event.Name)
					nodeModified, ok := root.index[nodePath]
					fmt.Println(event)
					if ok {
						nodeModified.Load(1)
					}
				}
			case <-done:
				return
			}
		}
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	done := make(chan interface{})
	go watch(watcher, done, node)

	children := []*FsNode{node}
	for len(children) > 0 {
		n := children[0]
		children = children[1:]
		if n.children != nil && len(n.children) > 0 {
			watcher.Add(n.path)
			children = append(children, n.children...)
		}
	}

	return done, nil
}
