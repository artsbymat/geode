package wikilink

import (
	"github.com/yuin/goldmark/ast"
)

var Kind = ast.NewNodeKind("WikiLink")

type Node struct {
	ast.BaseInline
	Target   []byte
	Fragment []byte
	Embed    bool
}

var _ ast.Node = (*Node)(nil)

func (n *Node) Kind() ast.NodeKind {
	return Kind
}

func (n *Node) Dump(src []byte, level int) {
	ast.DumpHelper(n, src, level, map[string]string{
		"Target": string(n.Target),
	}, nil)
}
