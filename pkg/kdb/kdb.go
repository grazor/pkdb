// Package kdb implements knowledge database logic
package kdb

import (
	"sync"
	"time"

	"github.com/grazor/pkdb/pkg/provider"
)

type KdbError struct {
	Inner   error
	Message string
}

func (err KdbError) Error() string {
	return err.Message
}

type KdbTree struct {
	Provider provider.Provider

	mu          sync.Mutex
	nodeCounter uint64

	Root   *KdbNode
	errors chan error
}

type KdbNode struct {
	ID          string
	NodeIndex   uint64
	Name        string
	HasChildren bool
	Attrs       map[string]interface{}

	Path string
	Size int64
	Time time.Time

	Parent   *KdbNode
	children map[string]*KdbNode

	Tree *KdbTree
}

func New(provider provider.Provider) *KdbTree {
	errors := make(chan error)
	root := &KdbNode{
		Name:        "root",
		Path:        "",
		HasChildren: true,
		NodeIndex:   1,
	}

	tree := &KdbTree{
		Provider:    provider,
		Root:        root,
		errors:      errors,
		nodeCounter: 1,
	}

	root.Tree = tree
	return tree
}

func (tree *KdbTree) Errors() <-chan error {
	return tree.errors
}

func (tree *KdbTree) Close() {
	close(tree.errors)
}

func (tree *KdbTree) nextIndex() uint64 {
	tree.mu.Lock()
	defer tree.mu.Unlock()
	tree.nodeCounter += 1
	index := tree.nodeCounter
	return index
}
