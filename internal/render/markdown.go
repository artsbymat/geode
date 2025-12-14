package render

import (
	"bytes"
	"geode/internal/content"
	"geode/internal/types"
	"geode/internal/utils"
	"os"
	"path/filepath"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"gopkg.in/yaml.v3"
)

func ParsingMarkdown(entries []content.FileEntry) []types.MetaMarkdown {
	pages := make([]types.MetaMarkdown, 0, len(entries))

	for _, entry := range entries {
		if !entry.IsMarkdown {
			continue
		}

		contentBytes, err := os.ReadFile(entry.Path)
		if err != nil {
			continue
		}

		frontmatter, body := extractFrontmatter(contentBytes)
		title := ExtractTitle(frontmatter, entry)
		link := ExtractPermalink(frontmatter, entry)

		wordCount := CountWords(string(body))
		readingTime := EstimateReadingTime(wordCount)

		htmlOut := renderToHTML(body)

		pages = append(pages, types.MetaMarkdown{
			Path:         entry.Path,
			RelativePath: entry.RelativePath,
			Link:         link,
			Title:        title,
			Frontmatter:  frontmatter,
			ReadingTime:  readingTime,
			WordCount:    wordCount,
			HTML:         htmlOut,
		})
	}

	return pages
}

func renderToHTML(source []byte) string {
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.Strikethrough,
			extension.Table,
			extension.TaskList,
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(html.WithUnsafe()),
	)

	var buf bytes.Buffer
	if err := md.Convert(source, &buf); err != nil {
		return ""
	}

	return buf.String()
}

func extractFrontmatter(src []byte) (map[string]any, []byte) {
	text := string(src)
	front := map[string]any{}

	if strings.HasPrefix(text, "---") {
		parts := strings.SplitN(text, "---", 3)
		if len(parts) >= 3 {
			yamlPart := strings.TrimSpace(parts[1])
			body := strings.TrimSpace(parts[2])

			yaml.Unmarshal([]byte(yamlPart), &front)
			return front, []byte(body)
		}
	}

	return front, src
}

func ExtractTitle(front map[string]any, entry content.FileEntry) string {
	if t, ok := front["title"]; ok {
		if s, ok := t.(string); ok {
			return s
		}
	}

	base := filepath.Base(entry.RelativePath)

	return strings.TrimSuffix(base, ".md")
}

func ExtractPermalink(front map[string]any, entry content.FileEntry) string {
	if t, ok := front["permalink"]; ok {
		if s, ok := t.(string); ok {
			return utils.PathToSlug(s)
		}
	}

	return strings.TrimSuffix(utils.PathToSlug(entry.RelativePath), ".md")
}

func CountWords(text string) int {
	words := strings.Fields(text)
	return len(words)
}

func EstimateReadingTime(wordCount int) int {
	const wordsPerMinute = 200
	min := wordCount / wordsPerMinute
	if min == 0 && wordCount > 0 {
		return 1
	}
	return min
}
