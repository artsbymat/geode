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

	if err := CopyThemeAssets(cfg); err != nil {
		return err
	}

	if err := CopyContentAssets(filtered, cfg); err != nil {
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

func CopyContentAssets(entries []content.FileEntry, cfg *config.Config) error {
	for _, entry := range entries {
		if !entry.IsAsset {
			continue
		}

		if shouldIgnoreAsset(entry.Path, cfg.IgnorePatterns) {
			continue
		}

		ext := filepath.Ext(entry.RelativePath)
		relPathNoExt := strings.TrimSuffix(entry.RelativePath, ext)
		relPathNoExt = strings.ReplaceAll(relPathNoExt, "\\", "/")

		normalizedPath := utils.PathToSlug(relPathNoExt) + ext
		destPath := filepath.Join("public", normalizedPath)

		if err := copyFile(entry.Path, destPath); err != nil {
			return fmt.Errorf("copy asset %s: %w", entry.RelativePath, err)
		}
	}

	return nil
}

func CopyThemeAssets(cfg *config.Config) error {
	srcDir := filepath.Join("themes", cfg.Theme, "assets")
	return copyDirRecursive(srcDir, "public")
}

func shouldIgnoreAsset(path string, patterns []string) bool {
	lowerPath := strings.ToLower(filepath.ToSlash(path))
	base := filepath.Base(lowerPath)

	for _, pattern := range patterns {
		pattern = strings.TrimSpace(strings.ToLower(pattern))
		if pattern == "" {
			continue
		}

		if strings.ContainsAny(pattern, "*?[") {
			if matched, _ := filepath.Match(pattern, base); matched {
				return true
			}
			continue
		}

		if strings.Contains(lowerPath, pattern) {
			return true
		}
	}
	return false
}

func copyDirRecursive(src, dest string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		targetPath := filepath.Join(dest, rel)

		if info.IsDir() {
			return os.MkdirAll(targetPath, 0o755)
		}

		return copyFile(path, targetPath)
	})
}

func copyFile(srcFile, destFile string) error {
	src, err := os.Open(srcFile)
	if err != nil {
		return err
	}
	defer src.Close()

	err = os.MkdirAll(filepath.Dir(destFile), 0o755)
	if err != nil {
		return err
	}

	dest, err := os.Create(destFile)
	if err != nil {
		return err
	}
	defer dest.Close()

	_, err = io.Copy(dest, src)
	if err != nil {
		return err
	}

	return dest.Sync()
}
