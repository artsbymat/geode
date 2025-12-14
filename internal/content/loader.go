package content

import (
	"bufio"
	"fmt"
	"geode/internal/config"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type FileEntry struct {
	Path         string
	RelativePath string
	Size         int64
	IsMarkdown   bool
	IsAsset      bool
}

func GetAllMarkdownAndAssets(srcDir string, cfg *config.Config) ([]FileEntry, error) {
	var entries []FileEntry

	err := filepath.WalkDir(srcDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if shouldSkip(srcDir, path, cfg.IgnorePatterns) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if d.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))

		isMarkdown := ext == ".md"
		isAsset := isAssetFile(ext)

		if !isMarkdown && !isAsset {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}

		entries = append(entries, FileEntry{
			Path:         path,
			RelativePath: rel,
			Size:         info.Size(),
			IsMarkdown:   isMarkdown,
			IsAsset:      isAsset,
		})

		return nil
	})

	if err != nil {
		return nil, err
	}

	return entries, nil
}

func shouldSkip(root, path string, patterns []string) bool {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}

	rel = filepath.ToSlash(rel)

	for _, p := range patterns {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}

		if matched, _ := filepath.Match(p, filepath.Base(rel)); matched {
			return true
		}

		if strings.Contains(rel, p) {
			return true
		}
	}

	return false
}

var assetExt = map[string]struct{}{
	".pdf": {}, ".csv": {},
	".mp3": {}, ".wav": {}, ".ogg": {},
	".mp4": {}, ".mov": {}, ".webm": {},
	".png": {}, ".jpg": {}, ".jpeg": {},
	".gif": {}, ".svg": {}, ".webp": {},
}

func isAssetFile(ext string) bool {
	_, ok := assetExt[ext]
	return ok
}

type RawNoteMeta struct {
	Title      string   `yaml:"title"`
	Publish    bool     `yaml:"publish"`
	Draft      bool     `yaml:"draft"`
	Permalink  string   `yaml:"permalink"`
	CssClasses []string `yaml:"cssClasses"`
	Aliases    []string `yaml:"aliases"`
	Tags       []string `yaml:"tags"`
}

func ReadFrontmatter(path string) (*RawNoteMeta, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024), 1024*1024)

	if !scanner.Scan() {
		return nil, nil
	}

	if strings.TrimSpace(scanner.Text()) != "---" {
		return nil, nil
	}

	var yamlLines []string

	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "---" {
			break
		}
		yamlLines = append(yamlLines, line)
	}

	var meta RawNoteMeta

	if err := yaml.Unmarshal([]byte(strings.Join(yamlLines, "\n")), &meta); err != nil {
		return nil, fmt.Errorf("frontmatter yaml error in %s: %w", path, err)
	}

	return &meta, nil
}

func FilterEntries(entries []FileEntry, cfg *config.Config) []FileEntry {
	out := make([]FileEntry, 0, len(entries))

	for _, e := range entries {
		if !e.IsMarkdown {
			out = append(out, e)
			continue
		}

		meta, err := ReadFrontmatter(e.Path)
		if err != nil {
			log.Printf("frontmatter error: %s (%v)", e.Path, err)
			continue
		}

		if meta == nil {
			meta = &RawNoteMeta{
				Draft: false,
			}
		}

		switch cfg.Build.Mode {

		case config.ModeDraft:
			if !meta.Draft {
				out = append(out, e)
			}

		case config.ModeExplicit:
			if meta.Publish {
				out = append(out, e)
			}
		}
	}

	return out
}
