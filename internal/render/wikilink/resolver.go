package wikilink

import (
	"path/filepath"
	"strings"
	"unicode"
)

type Resolver interface {
	ResolveWikilink(*Node) (destination []byte, err error)
}

type PageResolver struct {
	Pages         map[string]string
	ShortestPaths map[string]string
}

func (r PageResolver) ResolveWikilink(n *Node) ([]byte, error) {
	if len(n.Target) == 0 {
		return nil, nil
	}

	target := string(n.Target)
	target = strings.TrimSuffix(target, ".md\\")
	target = strings.Trim(target, "/")
	target = filepath.ToSlash(target)

	// Absolute Path
	if dest, ok := r.Pages[target]; ok {
		return withFragment(dest, n), nil
	}

	// Shortest Path
	base := filepath.Base(target)
	if dest, ok := r.ShortestPaths[base]; ok {
		return withFragment(dest, n), nil
	}

	return nil, nil
}

func withFragment(dest string, n *Node) []byte {
	if len(n.Fragment) > 0 {
		dest += "#" + transformHeadingID(string(n.Fragment))
	}
	return []byte(dest)
}

func transformHeadingID(text string) string {
	text = strings.TrimSpace(text)
	text = strings.TrimLeft(text, "#")
	text = strings.TrimSpace(text)

	var result strings.Builder

	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			result.WriteRune(unicode.ToLower(r))
		} else if unicode.IsSpace(r) || r == '-' || r == '_' {
			result.WriteRune('-')
		}
	}

	id := result.String()
	id = strings.Trim(id, "-")
	return id
}
