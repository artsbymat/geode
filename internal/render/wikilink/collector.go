package wikilink

import (
	"sync"
)

type LinkCollector struct {
	Renderer Resolver
	links    []CollectedLink
	mu       sync.Mutex
	hasDest  sync.Map
}

type CollectedLink struct {
	Title string
	URL   string
}

func NewLinkCollector(resolver Resolver) *LinkCollector {
	return &LinkCollector{
		Renderer: resolver,
		links:    make([]CollectedLink, 0),
	}
}

func (c *LinkCollector) CollectLink(n *Node, dest []byte, src []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()

	title := string(n.Target)
	if n.ChildCount() == 1 {
		labelBytes := nodeText(src, n.FirstChild())
		if len(labelBytes) > 0 {
			title = string(labelBytes)
		}
	}

	c.links = append(c.links, CollectedLink{
		Title: title,
		URL:   string(dest),
	})
}

func (c *LinkCollector) GetLinks() []CollectedLink {
	c.mu.Lock()
	defer c.mu.Unlock()
	result := make([]CollectedLink, len(c.links))
	copy(result, c.links)
	return result
}

func (c *LinkCollector) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.links = c.links[:0]
}
