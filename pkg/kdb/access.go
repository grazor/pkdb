package kdb

import (
	"fmt"
	"io"
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
