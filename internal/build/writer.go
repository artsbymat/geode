package build

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"geode/internal/config"
	"geode/internal/types"
	"geode/internal/utils"
)

type PageData struct {
	Name          template.HTML
	Suffix        template.HTML
	Title         template.HTML
	Tags          template.HTML
	WordCount     template.HTML
	ReadingTime   template.HTML
	Content       template.HTML
	Explorer      template.HTML
	Graph         template.HTML
	Toc           template.HTML
	OutgoingLinks template.HTML
	Backlinks     template.HTML
	Socials       template.HTML
	HasKatex      bool
	HasMermaid    bool
	HasTwitter    bool
	LiveReload    bool
	CSSClasses    []string
	Description   string
	Keywords      []string
	Date          template.HTML
}

type HTMLWriter struct {
	tmpl *template.Template
	cfg  *config.Config
}

func NewHTMLWriter(cfg *config.Config) (*HTMLWriter, error) {
	templatePath := filepath.Join("themes", cfg.Theme, "templates", "base.html")
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return nil, fmt.Errorf("parse template: %w", err)
	}

	return &HTMLWriter{
		tmpl: tmpl,
		cfg:  cfg,
	}, nil
}

func (w *HTMLWriter) Write(page types.MetaMarkdown, liveReload bool, fileTree *types.FileTree) error {
	var cleanPath string
	cleanPath = strings.TrimSuffix(utils.PathToSlug(page.RelativePath), ".md")
	outputPath := filepath.Join("public", cleanPath+".html")

	err := os.MkdirAll(filepath.Dir(outputPath), 0o755)
	if err != nil {
		return err
	}

	currentPageURL := page.Link
	if currentPageURL == "" {
		currentPageURL = "/" + utils.PathToSlug(page.RelativePath)
	}

	outgoingHTML := RenderLinkList(page.OutgoingLinks)
	backlinksHTML := RenderLinkList(page.Backlinks)
	tocHTML := RenderTOC(page.TableOfContents)
	tagsHTML := RenderTags(page.Tags)

	graphData := BuildGraph(page)
	graphHTML := RenderGraphView(graphData, currentPageURL)
	date := time.Now().Format("2006-01-02")

	// modified -> created -> now
	if v, ok := page.Frontmatter["modified"]; ok {
		switch t := v.(type) {
		case time.Time:
			date = t.Format("2006-01-02")
		case string:
			date = t
		}
	} else if v, ok := page.Frontmatter["created"]; ok {
		switch t := v.(type) {
		case time.Time:
			date = t.Format("2006-01-02")
		case string:
			date = t
		}
	}

	data := PageData{
		Name:          template.HTML(w.cfg.Site.Name),
		Suffix:        template.HTML(w.cfg.Site.Suffix),
		Title:         template.HTML(page.Title),
		Tags:          template.HTML(tagsHTML),
		WordCount:     template.HTML(strconv.Itoa(page.WordCount)),
		ReadingTime:   template.HTML(strconv.Itoa(page.ReadingTime)),
		Content:       template.HTML(page.HTML),
		Explorer:      template.HTML(RenderExplorer(fileTree)),
		Graph:         template.HTML(graphHTML),
		Toc:           template.HTML(tocHTML),
		OutgoingLinks: template.HTML(outgoingHTML),
		Backlinks:     template.HTML(backlinksHTML),
		Socials:       template.HTML(RenderSocials(w.cfg.Socials)),
		HasKatex:      page.HasKatex,
		HasMermaid:    page.HasMermaid,
		HasTwitter:    strings.Contains(page.HTML, `blockquote class="twitter-tweet"`),
		LiveReload:    liveReload,
		CSSClasses:    parseCSSClasses(page.Frontmatter),
		Description:   page.Description,
		Keywords:      parseKeywords(page.Frontmatter),
		Date:          template.HTML(date),
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	return w.tmpl.Execute(file, data)
}

func parseCSSClasses(front map[string]any) []string {
	v, ok := front["cssclasses"]
	if !ok || v == nil {
		return nil
	}

	var out []string
	switch vv := v.(type) {
	case []string:
		out = vv
	case []any:
		for _, it := range vv {
			if s, ok := it.(string); ok {
				out = append(out, s)
			}
		}
	case string:
		for _, s := range strings.Split(vv, " ") {
			s = strings.TrimSpace(s)
			if s != "" {
				out = append(out, s)
			}
		}
	}
	return out
}

func parseKeywords(front map[string]any) []string {
	v, ok := front["tags"]
	if !ok || v == nil {
		return nil
	}

	var out []string
	switch vv := v.(type) {
	case []string:
		out = vv
	case []any:
		for _, it := range vv {
			if s, ok := it.(string); ok {
				out = append(out, s)
			}
		}
	case string:
		for _, s := range strings.Split(vv, " ") {
			s = strings.TrimSpace(s)
			if s != "" {
				out = append(out, s)
			}
		}
	}
	return out
}
