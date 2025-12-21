package types

type Link struct {
	Title string
	URL   string
}

type TocItem struct {
	Level int
	Text  string
	ID    string
}

type MetaMarkdown struct {
	Path            string
	RelativePath    string
	Link            string
	Title           string
	Frontmatter     map[string]any
	Tags            []string
	ReadingTime     int
	WordCount       int
	HTML            string
	OutgoingLinks   []Link
	Backlinks       []Link
	TableOfContents []TocItem
	HasKatex        bool
}
