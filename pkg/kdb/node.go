package kdb

import (
	"fmt"
	"time"
)

type KdbNode struct {
	ID          string
	NodeIndex   uint64
	Name        string
	HasChildren bool
	Attrs       map[string]interface{}

	Path string
	Size int64
	Time time.Time
	MIME string

	Parent   *KdbNode
	children map[string]*KdbNode

	Tree *KdbTree
}

func (node *KdbNode) String() string {
	return fmt.Sprintf("Node %v (%v)", node.Path, node.Name)
}
