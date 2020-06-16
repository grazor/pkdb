package kdb

import (
	"fmt"
	"io"
	"path/filepath"
)

type kdbNodeWriter struct {
	node *KdbNode
	io.WriteCloser
}

func (w kdbNodeWriter) Close() error {
	err := w.WriteCloser.Close()
	if err != nil {
		return KdbError{
			Inner:   err,
			Message: fmt.Sprintf("unable to close after writing %v", w.node.Path),
		}
	}

	err = w.node.Reload()
	if err != nil {
		return KdbError{
			Inner:   err,
			Message: fmt.Sprintf("unable to reload node after writing %v", w.node.Path),
		}
	}

	return nil
}

func (node *KdbNode) Reader(off int64) (io.ReadCloser, error) {
	entry, err := node.Parent.Tree.Provider.Get(node.Path)
	if err != nil {
		return nil, KdbError{
			Inner:   err,
			Message: fmt.Sprintf("unable to get source node for %v", node.Path),
		}
	}

	reader, err := entry.Reader(off)
	if err != nil {
		return nil, KdbError{
			Inner:   err,
			Message: fmt.Sprintf("unable to open node for reading %v", node.Path),
		}
	}
	return reader, nil
}

func (node *KdbNode) Writer(off int64) (io.WriteCloser, error) {
	entry, err := node.Parent.Tree.Provider.Get(node.Path)
	if err != nil {
		return nil, KdbError{
			Inner:   err,
			Message: fmt.Sprintf("unable to get source node for %v", node.Path),
		}
	}

	writer, err := entry.Writer(off)
	if err != nil {
		return nil, KdbError{
			Inner:   err,
			Message: fmt.Sprintf("unable to open node for writing %v", node.Path),
		}
	}
	return kdbNodeWriter{node, writer}, nil
}

func (node *KdbNode) AddChild(name string, container bool) (newNode *KdbNode, err error) {
	if !node.HasChildren {
		return nil, KdbError{
			Message: fmt.Sprintf("unable to add child to leaf %v", node.Path),
		}
	}

	path := filepath.Join(node.Path, name)
	if _, ok := node.Child(name); ok {
		return nil, KdbError{
			Message: fmt.Sprintf("node already exists %v", path),
		}
	}

	providerParentNode, err := node.Tree.Provider.Get(node.Path)
	if err != nil {
		return nil, KdbError{
			Inner:   err,
			Message: fmt.Sprintf("unable to get provider parent node for %v", path),
		}

	}

	childProviderNode, err := providerParentNode.AddChild(name, container)
	if err != nil {
		return nil, KdbError{
			Inner:   err,
			Message: fmt.Sprintf("unable to create child for %v", path),
		}

	}

	kdbChildNode := nodeFromProvider(node, childProviderNode)
	return kdbChildNode, nil
}

func (node *KdbNode) Delete() error {
	children := node.Children()

	if len(children) > 0 {
		return KdbError{
			Message: fmt.Sprintf("can not delete node with children %s", node),
		}

	}

	providerNode, err := node.Tree.Provider.Get(node.Path)
	if err != nil {
		return KdbError{
			Inner:   err,
			Message: fmt.Sprintf("unable to get provider node %v", node.Path),
		}

	}

	err = providerNode.Delete()
	if err != nil {
		return KdbError{
			Inner:   err,
			Message: fmt.Sprintf("unable to get delete node %v", node.Path),
		}

	}

	delete(node.Parent.Children(), node.Name)
	return nil
}

func (node *KdbNode) Move(targetParent *KdbNode, name string) error {
	sourceProviderNode, err := node.Tree.Provider.Get(node.Path)
	if err != nil {
		return KdbError{
			Inner:   err,
			Message: fmt.Sprintf("unable to get sourceprovider node %v", node.Path),
		}

	}

	targetParentNode, err := node.Tree.Provider.Get(targetParent.Path)
	if err != nil {
		return KdbError{
			Inner:   err,
			Message: fmt.Sprintf("unable to get target parent node %v", targetParent.Path),
		}

	}

	if targetNode, ok := targetParent.Child(name); ok {
		// Replacing target node with a new one
		err = targetNode.Delete()
		if err != nil {
			return KdbError{
				Inner:   err,
				Message: fmt.Sprintf("unable to delete existing target node %v", targetParent.Path),
			}

		}
	}

	err = sourceProviderNode.Move(targetParentNode, name)
	if err != nil {
		return KdbError{
			Inner:   err,
			Message: "provider move failed",
		}
	}

	delete(node.Parent.children, node.Name)
	node.Parent = targetParent
	targetParent.Children()[name] = node

	node.Name = name
	node.Path = filepath.Join(targetParent.Path, node.Name)

	return nil
}
