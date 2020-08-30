package markdown

import (
	"io"
	"regexp"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
)

var _ renderer.Renderer = (*processingRenderer)(nil)

var (
	regexActionableIncomplete = regexp.MustCompile(`^\s*\[\s*\] `)
	regexActionableComplete   = regexp.MustCompile(`^\s*\[\s*[Xx]\s*\] `)
)

type processingRenderer struct {
	title                     string
	description               string
	isActionable, isCompleted bool
}

func newRenderer() *processingRenderer {
	return &processingRenderer{}
}

func innerMarkup(node ast.Node, source []byte) string {
	lines := node.Lines()
	firseSeg := lines.At(0)
	lastSeg := lines.At(lines.Len() - 1)
	seg := text.NewSegment(firseSeg.Start, lastSeg.Stop)
	return string(seg.Value(source))
}

func (r *processingRenderer) Render(w io.Writer, source []byte, n ast.Node) error {
	i := 0
	for n != nil && i < 2 {
		switch n.Kind() {
		case ast.KindDocument:
			n = n.FirstChild()
			continue
		case ast.KindHeading:
			h := n.(*ast.Heading)
			if i == 0 && h.Level <= 3 {
				r.title = innerMarkup(h, source)
			}
		case ast.KindParagraph:
			if i <= 1 && r.description == "" {
				p := n.(*ast.Paragraph)
				r.description = innerMarkup(p, source)
			}
		}
		n = n.NextSibling()
		i++
	}

	title := []byte(r.title)
	if regexActionableComplete.Match(title) {
		r.isActionable = true
		r.isCompleted = true
	} else if regexActionableIncomplete.Match(title) {
		r.isActionable = true
		r.isCompleted = false
	}

	return nil
}

func (r *processingRenderer) AddOptions(...renderer.Option) {

}
