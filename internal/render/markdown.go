package render

import (
	"bytes"
	"geode/internal/content"
	hashtag "geode/internal/render/tags"
	"geode/internal/render/wikilink"
	"geode/internal/types"
	"geode/internal/utils"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"gopkg.in/yaml.v3"
)

func ParsingMarkdown(entries []content.FileEntry) []types.MetaMarkdown {
	pages := make([]types.MetaMarkdown, 0, len(entries))
	urlToIndex := make(map[string]int, len(entries))
	pendingBacklinks := make(map[string][]types.Link)
	seenBacklinks := make(map[string]map[string]bool) // targetURL -> sourceURL -> seen

	resolver := buildResolver(entries)
	embedIndex := buildEmbedIndex(entries)

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

			htmlOut, outgoingLinks, toc, contentTags := renderToHTML(body, resolver, embedIndex, entry.Path)
			tags := mergeTags(parseFrontmatterTags(frontmatter), contentTags)

			page := types.MetaMarkdown{
				Path:            entry.Path,
				RelativePath:    entry.RelativePath,
				Link:            link,
				Title:           title,
				Frontmatter:     frontmatter,
				Tags:            tags,
				ReadingTime:     readingTime,
				WordCount:       wordCount,
				HTML:            htmlOut,
				OutgoingLinks:   outgoingLinks,
				TableOfContents: toc,
			}

			pages = append(pages, page)
			pageIndex := len(pages) - 1
			if link != "" {
				urlToIndex[link] = pageIndex
				if pending, ok := pendingBacklinks[link]; ok {
					pages[pageIndex].Backlinks = append(pages[pageIndex].Backlinks, pending...)
					delete(pendingBacklinks, link)
				}
			}

			sourceLink := types.Link{Title: title, URL: link}
			for _, out := range outgoingLinks {
				targetURL := out.URL
				if targetURL == "" || targetURL == link {
					continue
				}

				if _, ok := seenBacklinks[targetURL]; !ok {
					seenBacklinks[targetURL] = make(map[string]bool)
				}
				if seenBacklinks[targetURL][sourceLink.URL] {
					continue
				}
				seenBacklinks[targetURL][sourceLink.URL] = true

				if idx, ok := urlToIndex[targetURL]; ok {
					pages[idx].Backlinks = append(pages[idx].Backlinks, sourceLink)
				} else {
					pendingBacklinks[targetURL] = append(pendingBacklinks[targetURL], sourceLink)
				}
			}
		}
	}

	return pages
}

type embedResolver struct {
	Pages         map[string]string
	ShortestPaths map[string]string
}

func buildEmbedIndex(entries []content.FileEntry) embedResolver {
	pages := make(map[string]string)
	shortestPaths := make(map[string]string)
	baseNamePaths := make(map[string][]string)

	for _, entry := range entries {
		if !entry.IsMarkdown {
			continue
		}

		key := strings.TrimSuffix(entry.RelativePath, ".md")
		key = filepath.ToSlash(key)
		pages[key] = entry.Path

		base := filepath.Base(key)
		baseNamePaths[base] = append(baseNamePaths[base], key)
	}

	for base, paths := range baseNamePaths {
		shortestKey := paths[0]
		for _, key := range paths {
			if len(key) < len(shortestKey) {
				shortestKey = key
			}
		}
		shortestPaths[base] = pages[shortestKey]
	}

	return embedResolver{Pages: pages, ShortestPaths: shortestPaths}
}

func (r embedResolver) resolve(target string) (string, bool) {
	target = strings.TrimSuffix(target, ".md\\")
	target = strings.TrimSuffix(target, ".md")
	target = strings.Trim(target, "/")
	target = filepath.ToSlash(target)

	if dest, ok := r.Pages[target]; ok {
		return dest, true
	}

	base := filepath.Base(target)
	if dest, ok := r.ShortestPaths[base]; ok {
		return dest, true
	}

	return "", false
}

func expandMarkdownEmbeds(src []byte, r embedResolver, rootPath string) []byte {
	if len(src) == 0 {
		return src
	}

	type segment struct {
		b      []byte
		i      int
		onDone func()
	}

	const maxDepth = 20
	depth := 0

	includes := map[string]struct{}{rootPath: {}}
	stack := []segment{{b: src}}

	var out bytes.Buffer
	out.Grow(len(src))

	for len(stack) > 0 {
		seg := &stack[len(stack)-1]
		if seg.i >= len(seg.b) {
			if seg.onDone != nil {
				seg.onDone()
			}
			stack = stack[:len(stack)-1]
			continue
		}

		b := seg.b
		i := seg.i

		if b[i] != '!' || i+2 >= len(b) || b[i+1] != '[' || b[i+2] != '[' {
			_ = out.WriteByte(b[i])
			seg.i++
			continue
		}

		j := i + 3
		for j+1 < len(b) && !(b[j] == ']' && b[j+1] == ']') {
			j++
		}
		if j+1 >= len(b) {
			_ = out.WriteByte(b[i])
			seg.i++
			continue
		}

		inner := strings.TrimSpace(string(b[i+3 : j]))
		if inner == "" {
			_, _ = out.Write(b[i : j+2])
			seg.i = j + 2
			continue
		}

		literal := b[i : j+2]

		if k := strings.IndexByte(inner, '|'); k >= 0 {
			inner = inner[:k]
		}
		inner = strings.TrimSpace(inner)

		target := inner
		var fragmentID string
		if before, after, ok := strings.Cut(inner, "#"); ok {
			target = strings.TrimSpace(before)
			fragmentID = transformHeadingID(strings.TrimSpace(after))
		}
		if target == "" {
			_, _ = out.Write(literal)
			seg.i = j + 2
			continue
		}

		path, ok := r.resolve(target)
		if !ok {
			_, _ = out.Write(literal)
			seg.i = j + 2
			continue
		}

		if _, seen := includes[path]; seen || depth >= maxDepth {
			seg.i = j + 2
			continue
		}

		contentBytes, err := os.ReadFile(path)
		if err != nil {
			_, _ = out.Write(literal)
			seg.i = j + 2
			continue
		}

		_, body := extractFrontmatter(contentBytes)
		if fragmentID != "" {
			section, ok := extractMarkdownSection(body, fragmentID)
			if !ok {
				_, _ = out.Write(literal)
				seg.i = j + 2
				continue
			}
			body = section
		}
		includes[path] = struct{}{}
		depth++

		seg.i = j + 2
		stack = append(stack, segment{b: body, onDone: func() {
			delete(includes, path)
			depth--
		}})
	}

	return out.Bytes()
}

func extractMarkdownSection(body []byte, fragmentID string) ([]byte, bool) {
	if fragmentID == "" || len(body) == 0 {
		return nil, false
	}

	lines := strings.Split(string(body), "\n")

	startLine := -1
	startLevel := 0

	for i, line := range lines {
		level, text, ok := parseATXHeading(line)
		if !ok {
			continue
		}
		if transformHeadingID(text) == fragmentID {
			startLine = i
			startLevel = level
			break
		}
	}
	if startLine < 0 {
		return nil, false
	}

	endLine := len(lines)
	for i := startLine + 1; i < len(lines); i++ {
		level, _, ok := parseATXHeading(lines[i])
		if !ok {
			continue
		}
		if level <= startLevel {
			endLine = i
			break
		}
	}

	section := strings.Join(lines[startLine:endLine], "\n")
	return []byte(section), true
}

func parseATXHeading(line string) (level int, text string, ok bool) {
	trimmed := strings.TrimLeft(line, " \t")
	if trimmed == "" || trimmed[0] != '#' {
		return 0, "", false
	}

	level = 0
	for level < len(trimmed) && level < 6 && trimmed[level] == '#' {
		level++
	}
	if level == 0 {
		return 0, "", false
	}
	if level >= len(trimmed) || trimmed[level] != ' ' {
		return 0, "", false
	}

	text = strings.TrimSpace(trimmed[level:])
	text = strings.TrimRight(text, "#")
	text = strings.TrimSpace(text)
	if text == "" {
		return 0, "", false
	}

	return level, text, true
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

func renderToHTML(source []byte, resolver wikilink.Resolver, embed embedResolver, rootPath string) (string, []types.Link, []types.TocItem, []string) {
	collector := wikilink.NewLinkCollector(resolver)
	tagCollector := hashtag.NewCollector()
	toc := make([]types.TocItem, 0)
	tagResolver := hashtag.Resolver(tagLinkResolver{})

	source = expandMarkdownEmbeds(source, embed, rootPath)

	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.Strikethrough,
			extension.Table,
			extension.TaskList,
			&wikilink.Extender{
				Resolver:  resolver,
				Collector: collector,
			},
			&hashtag.Extender{
				Collector: tagCollector,
				Resolver:  tagResolver,
			},
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
			withHeadingShiftAndTOC(1, &toc),
		),
		goldmark.WithRendererOptions(html.WithUnsafe()),
	)

	var buf bytes.Buffer
	if err := md.Convert(source, &buf); err != nil {
		return "", nil, nil, nil
	}

	collectedLinks := collector.GetLinks()
	links := make([]types.Link, len(collectedLinks))
	for i, link := range collectedLinks {
		links[i] = types.Link{
			Title: link.Title,
			URL:   link.URL,
		}
	}

	return buf.String(), links, toc, tagCollector.Tags()
}

type tagLinkResolver struct{}

func (tagLinkResolver) ResolveHashtag(n *hashtag.Node) ([]byte, error) {
	tag := strings.TrimSpace(string(n.Tag))
	tag = strings.TrimPrefix(tag, "#")
	if tag == "" {
		return nil, nil
	}

	return []byte("/tags/" + escapeTagPath(tag)), nil
}

func escapeTagPath(tag string) string {
	escaped := url.PathEscape(tag)
	escaped = strings.ReplaceAll(escaped, "%2F", "/")
	escaped = strings.ReplaceAll(escaped, "%2f", "/")
	return escaped
}

func parseFrontmatterTags(front map[string]any) []string {
	v, ok := front["tags"]
	if !ok || v == nil {
		return nil
	}

	add := func(dst []string, s string) []string {
		s = strings.TrimSpace(s)
		s = strings.TrimPrefix(s, "#")
		if s == "" {
			return dst
		}
		return append(dst, s)
	}

	var out []string
	switch vv := v.(type) {
	case []string:
		for _, s := range vv {
			out = add(out, s)
		}
	case []any:
		for _, it := range vv {
			if s, ok := it.(string); ok {
				out = add(out, s)
			}
		}
	case string:
		// Some frontmatter uses: tags: tag1, tag2
		for s := range strings.SplitSeq(vv, ",") {
			out = add(out, s)
		}
	}

	return out
}

func mergeTags(a, b []string) []string {
	if len(a) == 0 {
		return b
	}
	if len(b) == 0 {
		return a
	}

	seen := make(map[string]struct{}, len(a)+len(b))
	out := make([]string, 0, len(a)+len(b))
	for _, s := range a {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	for _, s := range b {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}

	sort.Strings(out)
	return out
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
			url := strings.TrimSuffix(utils.PathToSlug(s), ".md")
			url = strings.TrimSpace(url)
			if url == "" {
				return ""
			}
			if !strings.HasPrefix(url, "/") {
				url = "/" + url
			}
			return url
		}
	}

	url := strings.TrimSuffix(utils.PathToSlug(entry.RelativePath), ".md")
	url = strings.TrimSpace(url)
	if url == "" {
		return ""
	}
	if !strings.HasPrefix(url, "/") {
		url = "/" + url
	}
	return url
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
