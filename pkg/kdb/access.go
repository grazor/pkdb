package kdb

import (
	"fmt"
	"io"
	"path/filepath"
)

func (node *KdbNode) Open() (io.ReadWriteCloser, error) {
	entry, err := node.Parent.Tree.Provider.Get(node.Path)
	if err != nil {
		return nil, KdbError{
			Inner:   err,
			Message: fmt.Sprintf("unable to get source node for %v", node.Path),
		}
	}

	rwcloser, err := entry.Open()
	if err != nil {
		return nil, KdbError{
			Inner:   err,
			Message: fmt.Sprintf("unable to open node for read %v", node.Path),
		}
	}
	return rwcloser, nil
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
