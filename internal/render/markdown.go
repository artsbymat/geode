package render

import (
	"bytes"
	"geode/internal/content"
	"geode/internal/render/wikilink"
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

	resolver := buildResolver(entries)

	for _, entry := range entries {
		if entry.IsAsset {
			continue
		} else if entry.IsMarkdown {

			contentBytes, err := os.ReadFile(entry.Path)
			if err != nil {
				continue
			}

			frontmatter, body := extractFrontmatter(contentBytes)
			title := ExtractTitle(frontmatter, entry)
			link := ExtractPermalink(frontmatter, entry)

			wordCount := CountWords(string(body))
			readingTime := EstimateReadingTime(wordCount)

			htmlOut := renderToHTML(body, resolver)

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
	}

	return pages
}

func buildResolver(entries []content.FileEntry) wikilink.Resolver {
	pages := make(map[string]string)
	shortestPaths := make(map[string]string)
	baseNamePaths := make(map[string][]string)

	for _, entry := range entries {
		key := strings.TrimSuffix(entry.RelativePath, ".md")
		key = filepath.ToSlash(key)

		link := utils.PathToSlug(entry.RelativePath)
		link = "/" + strings.TrimSuffix(link, ".md")

		pages[key] = link

		base := filepath.Base(key)
		baseNamePaths[base] = append(baseNamePaths[base], key)
	}

	for base, paths := range baseNamePaths {
		shortestPath := paths[0]
		for _, path := range paths {
			if len(path) < len(shortestPath) {
				shortestPath = path
			}
		}
		shortestPaths[base] = pages[shortestPath]
	}

	return wikilink.PageResolver{
		Pages:         pages,
		ShortestPaths: shortestPaths,
	}
}

func renderToHTML(source []byte, resolver wikilink.Resolver) string {
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.Strikethrough,
			extension.Table,
			extension.TaskList,
			&wikilink.Extender{
				Resolver: resolver,
			},
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
