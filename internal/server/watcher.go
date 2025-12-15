package server

import (
	"fmt"
	"geode/internal/build"
	"geode/internal/config"
	"geode/internal/content"
	"geode/internal/render"
	"geode/internal/utils"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

func WatchAndRebuild(contentDir string, cfg *config.Config) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	themesPath := filepath.Join("themes", cfg.Theme)

	if err := watchRecursive(watcher, contentDir); err != nil {
		log.Fatal(err)
	}
	if err := watchRecursive(watcher, themesPath); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Watching for changes...")

	var (
		debounce *time.Timer
		mu       sync.Mutex
	)

	for {
		select {
		case e := <-watcher.Events:
			if e.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove|fsnotify.Rename) == 0 {
				continue
			}

			name := filepath.Base(e.Name)

			// Ignore temp / editor files
			if strings.HasSuffix(name, "~") ||
				strings.HasSuffix(name, ".swp") ||
				strings.Contains(name, "4913") ||
				strings.HasPrefix(name, ".#") {
				continue
			}

			if e.Op&fsnotify.Create != 0 {
				if info, err := os.Stat(e.Name); err == nil && info.IsDir() {
					_ = watchRecursive(watcher, e.Name)
				}
			}

			event := e

			mu.Lock()
			if debounce != nil {
				debounce.Stop()
			}

			debounce = time.AfterFunc(200*time.Millisecond, func() {
				fmt.Println("Changed:", event.Name)

				if err := Rebuild(contentDir, cfg, true); err != nil {
					log.Println("Rebuild error:", err)
					return
				}

				BroadcastReload()
			})
			mu.Unlock()

		case err := <-watcher.Errors:
			log.Println("Watcher error:", err)
		}
	}
}

func watchRecursive(w *fsnotify.Watcher, root string) error {
	return filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			if err := w.Add(path); err != nil {
				return err
			}
		}
		return nil
	})
}

func Rebuild(dir string, cfg *config.Config, live bool) error {
	if err := CleanPublicDir(); err != nil {
		return fmt.Errorf("clean public dir: %w", err)
	}

	entries, err := content.GetAllMarkdownAndAssets(dir, cfg)
	if err != nil {
		return err
	}

	filtered := content.FilterEntries(entries, cfg)

	pages := render.ParsingMarkdown(filtered)

	fileTree := render.BuildFileTree(pages)

	writer, err := build.NewHTMLWriter(cfg)
	if err != nil {
		return fmt.Errorf("init html writer: %w", err)
	}

	for _, page := range pages {
		if err := writer.Write(page, live, fileTree); err != nil {
			return fmt.Errorf("write html %s: %w", page.RelativePath, err)
		}
	}

	// TODO: Build Tags Pages
	// TODO: Build default directory pages
	// TODO: Build 404 Pages

	if err := CopyAssets(filtered, cfg); err != nil {
		return err
	}

	if live {
		fmt.Println("Site rebuilt.")
	} else {
		fmt.Println("Build completed.")
	}

	return nil
}

func CleanPublicDir() error {
	err := os.RemoveAll("public")
	if err != nil {
		return err
	}

	return os.MkdirAll("public", 0o755)
}

func CopyAssets(entries []content.FileEntry, cfg *config.Config) error {
	for _, entry := range entries {
		if !entry.IsAsset {
			continue
		}

		if shouldSkipAsset(entry.Path, cfg.IgnorePatterns) {
			continue
		}

		ext := filepath.Ext(entry.RelativePath)
		relPathNoExt := strings.TrimSuffix(entry.RelativePath, ext)
		relPathNoExt = strings.ReplaceAll(relPathNoExt, "\\", "/")

		normalizedPath := utils.PathToSlug(relPathNoExt) + ext
		destPath := filepath.Join("public", normalizedPath)

		if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
			return fmt.Errorf("create asset dir: %w", err)
		}

		if err := copyAssetFile(entry.Path, destPath); err != nil {
			return fmt.Errorf("copy asset %s: %w", entry.RelativePath, err)
		}
	}

	return nil
}

func shouldSkipAsset(path string, patterns []string) bool {
	lowerPath := strings.ToLower(filepath.ToSlash(path))
	base := filepath.Base(lowerPath)

	for _, raw := range patterns {
		p := strings.TrimSpace(strings.ToLower(raw))
		if p == "" {
			continue
		}

		if strings.ContainsAny(p, "*?[") {
			if ok, _ := filepath.Match(p, base); ok {
				return true
			}
			continue
		}

		if strings.Contains(lowerPath, p) {
			return true
		}
	}
	return false
}

func copyAssetFile(src, dest string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	destFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, srcFile); err != nil {
		return err
	}

	return destFile.Sync()
}
