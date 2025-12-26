package build

import (
	"html/template"
	"net/url"
	"strings"
)

func RenderTags(tags []string) string {
	if len(tags) == 0 {
		return ""
	}

	var b strings.Builder
	first := true
	for _, t := range tags {
		t = strings.TrimSpace(t)
		t = strings.TrimPrefix(t, "#")
		if t == "" {
			continue
		}
		if !first {
			b.WriteString("\n")
		}
		first = false
		href := "/tags/" + escapeTagPath(t)
		b.WriteString(`<li class="tag-item"><a class="tag" href="`)
		b.WriteString(template.HTMLEscapeString(href))
		b.WriteString(`">#`)
		b.WriteString(template.HTMLEscapeString(t))
		b.WriteString(`</a></li>`)
	}

	return b.String()
}

func escapeTagPath(tag string) string {
	escaped := url.PathEscape(tag)
	escaped = strings.ReplaceAll(escaped, "%2F", "/")
	escaped = strings.ReplaceAll(escaped, "%2f", "/")
	return escaped
}
