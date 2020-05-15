package filesystem

import (
	"github.com/grazor/pkdb/pkg/kdb"
	"os"
	"path/filepath"
	"sync"
)

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
			childNode := &FsNode{
				DataNode: &kdb.DataNode{
					Name:   content.Name(),
					Path:   filepath.Join(scanNode.Path, content.Name()),
					IsLeaf: !content.IsDir(),
				},
				path:   fullPath,
				parent: scanNode,
				index:  scanNode.index,
			}
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
