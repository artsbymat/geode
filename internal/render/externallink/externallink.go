package externallink

import (
	"geode/internal/build"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

type Extender struct{}

func (e *Extender) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithASTTransformers(
			util.Prioritized(&Transformer{}, 1000),
		),
	)
	m.Renderer().AddOptions(
		renderer.WithNodeRenderers(
			util.Prioritized(&ExternalLinkRenderer{}, 100),
		),
	)
}

type Transformer struct{}

func (t *Transformer) Transform(node *ast.Document, reader text.Reader, pc parser.Context) {
	ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		if n.Kind() == ast.KindLink {
			link := n.(*ast.Link)
			dest := string(link.Destination)

			if isExternalLink(dest) {
				link.SetAttributeString("data-external", true)
			}
		}

		return ast.WalkContinue, nil
	})
}

func isExternalLink(dest string) bool {
	if dest == "" {
		return false
	}

	if strings.HasPrefix(dest, "http://") || strings.HasPrefix(dest, "https://") {
		return true
	}

	if strings.HasPrefix(dest, "//") {
		return true
	}

	return false
}

type ExternalLinkRenderer struct {
	html.Config
}

func (r *ExternalLinkRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindLink, r.renderLink)
}

func (r *ExternalLinkRenderer) renderLink(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.Link)

	if entering {
		_, _ = w.WriteString("<a href=\"")
		_, _ = w.Write(util.URLEscape(n.Destination, true))
		_, _ = w.WriteString("\"")

		if n.Title != nil {
			_, _ = w.WriteString(` title="`)
			_, _ = w.Write(n.Title)
			_, _ = w.WriteString(`"`)
		}

		isExternal, _ := n.AttributeString("data-external")
		if isExternal != nil {
			if ext, ok := isExternal.(bool); ok && ext {
				_, _ = w.WriteString(` class="external-link" target="_blank" rel="noopener noreferrer"`)
			}
		}

		_, _ = w.WriteString(">")
	} else {
		isExternal, _ := n.AttributeString("data-external")
		if isExternal != nil {
			if ext, ok := isExternal.(bool); ok && ext {
				_, _ = w.WriteString(build.ExternalLinkIcon)
			}
		}

		_, _ = w.WriteString("</a>")
	}

	return ast.WalkContinue, nil
}
