package kdb

import "fmt"

func (node *KdbNode) Meta() (map[string]interface{}, error) {
	entry, err := node.Parent.Tree.Provider.Get(node.Path)
	if err != nil {
		return nil, KdbError{
			Inner:   err,
			Message: fmt.Sprintf("unable to get source node for %v", node.Path),
		}
	}

	meta, err := entry.Meta()
	if err != nil {
		return nil, KdbError{
			Inner:   err,
			Message: fmt.Sprintf("unable to get metadata for %v", node.Path),
		}
	}
	return meta, nil
}

func (node *KdbNode) SetMeta(meta map[string]interface{}) error {
	entry, err := node.Parent.Tree.Provider.Get(node.Path)
	if err != nil {
		return KdbError{
			Inner:   err,
			Message: fmt.Sprintf("unable to get source node for %v", node.Path),
		}
	}

	err = entry.SetMeta(meta)
	if err != nil {
		return KdbError{
			Inner:   err,
			Message: fmt.Sprintf("unable to get metadata for %v", node.Path),
		}
	}
	return nil
}
