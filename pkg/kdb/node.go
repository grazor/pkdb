package kdb

import "sync"

func (node *kdbNode) Children() map[string]*kdbNode {
	entry, err := parent.Tree.Provider.Get(parent.Path)
}

func (node *kdbNode) Load(depth int) error {
	if depth == 0 {
		return nil
	}

	var scan func(parent *kdbNode, depth int, errors chan<- error, wg *sync.WaitGroup)
	scan = func(parent *kdbNode, depth int, errors chan<- error, wg *sync.WaitGroup) {
		defer wg.Done()
		if depth == 0 {
			return
		}

		entry, err := parent.Tree.Provider.Get(parent.Path)
		if err != nil {
			errors <- err
			return
		}

		children, err := entry.Children()
		if err != nil {
			errors <- err
			return
		}

		parent.children = make(map[string]*kdbNode)
		for _, child := range children {
			childNode := &kdbNode{
				ID:     child.ID(),
				Name:   child.Name(),
				Path:   child.Path(),
				Attrs:  child.Attrs(),
				Parent: parent,
				Tree:   parent.Tree,
			}
			parent.children[childNode.Name] = childNode

			if child.HasChildren() {
				wg.Add(1)
				go scan(childNode, depth-1, errors, wg)
			}
		}
	}

	errors := make(chan error)
	var wg *sync.WaitGroup
	wg.Add(1)
	go scan(node, depth-1, errors, wg)
	wg.Wait()

	return nil
}
