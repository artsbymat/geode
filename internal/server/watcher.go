package server

import (
	"fmt"
	"geode/internal/build"
	"geode/internal/config"
	"geode/internal/content"
	"geode/internal/render"
	"os"
)

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

	writer, err := build.NewHTMLWriter(cfg)
	if err != nil {
		return fmt.Errorf("init html writer: %w", err)
	}

	for _, page := range pages {
		if err := writer.Write(page, live); err != nil {
			return fmt.Errorf("write html %s: %w", page.RelativePath, err)
		}
	}

	// TODO: Build Tags Pages
	// TODO: Build default directory pages
	// TODO: Build 404 Pages
	// TODO: Copy static files

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
