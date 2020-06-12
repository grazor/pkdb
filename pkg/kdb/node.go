package kdb

import (
	"fmt"
)

func (node *KdbNode) String() string {
	return fmt.Sprintf("Node %v (%v)", node.Path, node.Name)
}

