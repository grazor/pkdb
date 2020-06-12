package kdb

import "fmt"

func (node *KdbNode) Read(p []byte) (n int, err error) {
	entry, err := node.Parent.Tree.Provider.Get(node.Path)
	if err != nil {
		return 0, KdbError{
			Inner:   err,
			Message: fmt.Sprintf("unable to get source node for %v", node.Path),
		}
	}

	return entry.Read(p)
}
