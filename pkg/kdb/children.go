package kdb

import (
	"fmt"
	"sync"
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
			childNode := &KdbNode{
				ID:          child.ID(),
				Name:        child.Name(),
				Path:        child.Path(),
				Size:        child.Size(),
				Time:        child.Time(),
				HasChildren: child.HasChildren(),
				Attrs:       child.Attrs(),
				Parent:      parent,
				Tree:        parent.Tree,
			}
			parent.children[childNode.Name] = childNode

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

