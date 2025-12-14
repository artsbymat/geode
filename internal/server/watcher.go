package server

import (
	"fmt"
	"geode/internal/config"
	"geode/internal/content"
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
	fmt.Println(len(filtered))

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
