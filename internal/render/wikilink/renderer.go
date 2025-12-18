package wikilink

import (
	"bytes"
	"fmt"
	"io"
	"path/filepath"
	"sync"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

type Renderer struct {
	Resolver Resolver
	hasDest  sync.Map
}

func (r *Renderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(Kind, r.Render)
}

func (r *Renderer) Render(w util.BufWriter, src []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n, ok := node.(*Node)
	if !ok {
		return ast.WalkStop, fmt.Errorf("unexpected node %T, expected *wikilink.Node", node)
	}

	if entering {
		return r.enter(w, n, src)
	}

	r.exit(w, n)
	return ast.WalkContinue, nil
}

func (r *Renderer) enter(w util.BufWriter, n *Node, src []byte) (ast.WalkStatus, error) {
	dest, err := r.Resolver.ResolveWikilink(n)
	if err != nil {
		return ast.WalkStop, fmt.Errorf("resolve %q: %w", n.Target, err)
	}
	if len(dest) == 0 {
		return ast.WalkContinue, nil
	}

	img := resolveAsImage(n)
	if !img {
		r.hasDest.Store(n, struct{}{})
		_, _ = w.WriteString(`<a href="`)
		_, _ = w.Write(util.URLEscape(dest, true /* resolve references */))
		_, _ = w.WriteString(`">`)
		return ast.WalkContinue, nil
	}

	_, _ = w.WriteString(`<img src="`)
	_, _ = w.Write(util.URLEscape(dest, true /* resolve references */))
	if n.ChildCount() == 1 {
		label := nodeText(src, n.FirstChild())
		if !bytes.Equal(label, n.Target) {
			_, _ = w.WriteString(`" alt="`)
			_, _ = w.Write(util.EscapeHTML(label))
		}
	}
	_, _ = w.WriteString(`">`)
	return ast.WalkSkipChildren, nil
}

func (r *Renderer) exit(w util.BufWriter, n *Node) {
	if _, ok := r.hasDest.LoadAndDelete(n); ok {
		_, _ = w.WriteString("</a>")
	}
}

func resolveAsImage(n *Node) bool {
	if !n.Embed {
		return false
	}

	filename := string(n.Target)
	switch ext := filepath.Ext(filename); ext {
	case ".apng", ".avif", ".gif", ".jpg", ".jpeg", ".jfif", ".pjpeg", ".pjp", ".png", ".svg", ".webp":
		return true
	default:
		return false
	}
}

func nodeText(src []byte, n ast.Node) []byte {
	var buf bytes.Buffer
	writeNodeText(src, &buf, n)
	return buf.Bytes()
}

func writeNodeText(src []byte, dst io.Writer, n ast.Node) {
	switch n := n.(type) {
	case *ast.Text:
		_, _ = dst.Write(n.Segment.Value(src))
	case *ast.String:
		_, _ = dst.Write(n.Value)
	default:
		for c := n.FirstChild(); c != nil; c = c.NextSibling() {
			writeNodeText(src, dst, c)
		}
	}
}
