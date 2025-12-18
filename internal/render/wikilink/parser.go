package wikilink

import (
	"bytes"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

type Parser struct{}

var _ parser.InlineParser = (*Parser)(nil)

var (
	_open      = []byte("[[")
	_embedOpen = []byte("![[")
	_pipe      = []byte{'|'}
	_hash      = []byte{'#'}
	_close     = []byte("]]")
)

func (p *Parser) Trigger() []byte {
	return []byte{'!', '['}
}

func (p *Parser) Parse(_ ast.Node, block text.Reader, _ parser.Context) ast.Node {
	line, seg := block.PeekLine()
	stop := bytes.Index(line, _close)
	if stop < 0 {
		return nil
	}

	var embed bool

	switch {
	case bytes.HasPrefix(line, _open):
		seg = text.NewSegment(seg.Start+len(_open), seg.Start+stop)
	case bytes.HasPrefix(line, _embedOpen):
		embed = true
		seg = text.NewSegment(seg.Start+len(_embedOpen), seg.Start+stop)
	default:
		return nil
	}

	n := &Node{Target: block.Value(seg), Embed: embed}
	if idx := bytes.Index(n.Target, _pipe); idx >= 0 {
		n.Target = n.Target[:idx]                // [[ ... |
		seg = seg.WithStart(seg.Start + idx + 1) // | ... ]]
	}

	if len(n.Target) == 0 || seg.Len() == 0 {
		return nil
	}

	if idx := bytes.LastIndex(n.Target, _hash); idx >= 0 {
		n.Fragment = n.Target[idx+1:] // Foo#Bar => Bar
		n.Target = n.Target[:idx]     // Foo#Bar => Foo
	}

	n.AppendChild(n, ast.NewTextSegment(seg))
	block.Advance(stop + 2)
	return n
}
