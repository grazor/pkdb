package filesystem

import (
	"fmt"
	"strings"
)

func (node *FsNode) String() string {
	return fmt.Sprintf("%v: %v\n", node.path, node.Name)
}

func (node *FsNode) GetTree() string {
	type nodeStruct struct {
		n *FsNode
		d int
	}
	children := []nodeStruct{{node, 0}}
	var info string

	for len(children) > 0 {
		n, d := children[0].n, children[0].d
		info += fmt.Sprintf("\n%v-%v (%v)", strings.Repeat(" |", d+1), n.Name, n.path)

		children = children[1:]
		if len(n.children) > 0 {
			for _, child := range n.children {
				children = append([]nodeStruct{{child, d + 1}}, children...)
			}
		}
	}

	return fmt.Sprintln(info)
}

