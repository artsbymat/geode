package main

import (
	"flag"
	"fmt"
	"geode/internal/config"
	"geode/internal/server"
	"log"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {

	case "serve":
		runServe(os.Args[2:])

	case "build":
		runBuild(os.Args[2:])

	default:
		fmt.Println("Unknown command:", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func runServe(args []string) {
	serveCmd := flag.NewFlagSet("serve", flag.ExitOnError)
	port := serveCmd.Int("port", 3001, "application port")
	contentDir := serveCmd.String("dir", "content", "content directory")

	serveCmd.Parse(args)

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Serving %s at http://localhost:%d\n", *contentDir, *port)

	// build once
	err = server.Rebuild(*contentDir, cfg, true)
	if err != nil {
		log.Fatal(err)
	}

	go server.WatchAndRebuild(*contentDir, cfg)
	server.ServePublic(*port)
}

func runBuild(args []string) {
	buildCmd := flag.NewFlagSet("build", flag.ExitOnError)
	contentDir := buildCmd.String("dir", "content", "content directory")

	buildCmd.Parse(args)

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Building from:", *contentDir)
	err = server.Rebuild(*contentDir, cfg, false)
	if err != nil {
		log.Fatal(err)
	}
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  geode build [flags]")
	fmt.Println("  geode serve [flags]")
}
