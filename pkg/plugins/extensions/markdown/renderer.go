package markdown

import (
	"io"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
)

var _ renderer.Renderer = (*processingRenderer)(nil)

type processingRenderer struct {
	title                    string
	description              string
	isActioable, isCompleted bool
}

func newRenderer() *processingRenderer {
	return &processingRenderer{}
}

func (r *processingRenderer) Render(w io.Writer, source []byte, n ast.Node) error {
	return nil
}

func (r *processingRenderer) AddOptions(...renderer.Option) {

}
