package wikilink

import (
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

type Extender struct {
	Resolver  Resolver
	Collector *LinkCollector
}

func (e *Extender) Extend(md goldmark.Markdown) {
	md.Parser().AddOptions(
		parser.WithInlineParsers(
			util.Prioritized(&Parser{}, 199),
		),
	)

	md.Renderer().AddOptions(
		renderer.WithNodeRenderers(
			util.Prioritized(&Renderer{
				Resolver:  e.Resolver,
				Collector: e.Collector,
			}, 199),
		),
	)
}
