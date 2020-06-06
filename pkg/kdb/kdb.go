package kdb

import "github.com/grazor/pkdb/pkg/provider"

type KdbError struct {
	Inner   error
	Message string
}

func (err KdbError) Error() string {
	return err.Message
}

type kdbTree struct {
	Provider *provider.Provider

	Root   *kdbNode
	errors chan<- KdbError
}

type kdbNode struct {
	ID    string
	Name  string
	Path  string
	Attrs map[string]interface{}

	Parent *kdbNode

	children map[string]*kdbNode

	Tree *kdbTree
}

func New(provider *provider.Provider) (*kdbTree, <-chan KdbError) {
	errors = make(chan KdbError)
	root := &kdbNode{
		Name: "root",
		Path: "",
	}

	tree := &kdbTree{
		Provider: provider,
		Root:     root,
		errors:   errors,
	}

	root.Tree = tree
	return tree, errors
}
