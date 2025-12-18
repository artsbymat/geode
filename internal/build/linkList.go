package build

import (
	"geode/internal/types"
	"html/template"
	"sort"
	"strings"
)

func RenderLinkList(links []types.Link) string {
	if len(links) == 0 {
		return ""
	}

	sorted := make([]types.Link, 0, len(links))
	for _, l := range links {
		if strings.TrimSpace(l.URL) == "" {
			continue
		}
		sorted = append(sorted, l)
	}
	if len(sorted) == 0 {
		return ""
	}

	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Title == sorted[j].Title {
			return sorted[i].URL < sorted[j].URL
		}
		return sorted[i].Title < sorted[j].Title
	})

	var b strings.Builder
	b.WriteString(`<ul>`)
	for _, l := range sorted {
		label := strings.TrimSpace(l.Title)
		if label == "" {
			label = l.URL
		}
		b.WriteString(`<li><a href="`)
		b.WriteString(template.HTMLEscapeString(l.URL))
		b.WriteString(`">`)
		b.WriteString(template.HTMLEscapeString(label))
		b.WriteString(`</a></li>`)
	}
	b.WriteString(`</ul>`)
	return b.String()
}
