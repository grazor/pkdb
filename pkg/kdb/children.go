package kdb

import (
	"fmt"
	"sync"

	"github.com/grazor/pkdb/pkg/provider"
)

func (node *KdbNode) Children() map[string]*KdbNode {
	if !node.HasChildren {
		return nil
	}
	if node.children == nil {
		node.Load(1)
	}
	return node.children
}

func (node *KdbNode) Child(name string) (*KdbNode, bool) {
	n, ok := node.Children()[name]
	return n, ok
}

func (node *KdbNode) Load(depth int) {
	var scan func(parent *KdbNode, depth int, wg *sync.WaitGroup)
	scan = func(parent *KdbNode, depth int, wg *sync.WaitGroup) {
		defer wg.Done()
		if depth == 0 {
			return
		}

		entry, err := parent.Tree.Provider.Get(parent.Path)
		if err != nil {
			select {
			case parent.Tree.errors <- KdbError{
				Inner:   err,
				Message: fmt.Sprintf("could not get %v node", parent.Path),
			}:
			default:
			}

			return
		}

		children, err := entry.Children()
		if err != nil {
			select {
			case parent.Tree.errors <- KdbError{
				Inner:   err,
				Message: fmt.Sprintf("could not get %v node children", parent.Path),
			}:
			default:
			}
			return
		}

		parent.children = make(map[string]*KdbNode)
		for _, child := range children {
			childNode := nodeFromProvider(parent, child)
			if child.HasChildren() {
				wg.Add(1)
				go scan(childNode, depth-1, wg)
			}
		}
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go scan(node, depth, &wg)
	wg.Wait()
}

func nodeFromProvider(parent *KdbNode, entry provider.Entry) *KdbNode {
	node := &KdbNode{
		ID:          entry.ID(),
		Name:        entry.Name(),
		Path:        entry.Path(),
		Size:        entry.Size(),
		Time:        entry.Time(),
		HasChildren: entry.HasChildren(),
		Attrs:       entry.Attrs(),
	}

	if parent != nil {
		node.Parent = parent
		node.Tree = parent.Tree
		parent.children[node.Name] = node
	}

	return node
}
