// Package kdb implements knowledge database logic
package kdb

import (
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

	Root   *KdbNode
	errors chan error
}

type KdbNode struct {
	ID          string
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
	}

	tree := &KdbTree{
		Provider: provider,
		Root:     root,
		errors:   errors,
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
