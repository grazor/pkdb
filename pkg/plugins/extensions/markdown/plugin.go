// Package markdown implements pkdb plugin for metadata extraction from markdown files.
package markdown

import (
	"context"
	"io/ioutil"

	"github.com/grazor/pkdb/pkg/kdb"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
)

var _ kdb.KdbPlugin = (*markdownPlugin)(nil)
var _ kdb.KdbNodeUpdater = (*markdownPlugin)(nil)

type markdownPlugin struct {
	tree *kdb.KdbTree
}

func New() *markdownPlugin {
	return &markdownPlugin{}
}

func (p *markdownPlugin) Init(tree *kdb.KdbTree) error {
	p.tree = tree
	return nil
}

func (p *markdownPlugin) ID() string {
	return "markdown"
}

func (p *markdownPlugin) NodeUpdated(ctx context.Context, node *kdb.KdbNode, data []byte) {
	if node.MIME != "text/markdown" {
		return
	}

	r := newRenderer()
	md := goldmark.New(goldmark.WithParserOptions(parser.WithAttribute()), goldmark.WithRenderer(r))
	err := md.Convert(data, ioutil.Discard)
	if err != nil {
		p.tree.Errors() <- kdb.KdbError{
			Inner:   err,
			Message: "unable to process markdown data",
		}
		return
	}

	attributes := make(map[string]interface{})
	if r.title != "" {
		attributes["title"] = r.title
	}
	if r.description != "" {
		attributes["description"] = r.description
	}
	attributes["isActionable"] = r.isActionable
	if r.isActionable {
		attributes["isCompleted"] = r.isCompleted
	}

	err = node.UpdateMeta(attributes)
	if err != nil {
		p.tree.Errors() <- kdb.KdbError{
			Inner:   err,
			Message: "unable to process markdown data",
		}
		return
	}
}
