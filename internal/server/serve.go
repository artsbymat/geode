package server

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func ServePublic(port int) {
	fs := http.FileServer(http.Dir("public"))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := filepath.Clean(r.URL.Path)

		if path == "/" {
			http.ServeFile(w, r, "public/index.html")
			return
		}

		rel := strings.TrimPrefix(path, "/")
		htmlPath := filepath.Join("public", rel) + ".html"

		if fi, err := os.Stat(htmlPath); err == nil && !fi.IsDir() {
			http.ServeFile(w, r, htmlPath)
			return
		}

		fs.ServeHTTP(w, r)
	})

	http.HandleFunc("/_reload", sseHandler)

	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(port), nil))
}

var sseClients = make(map[chan string]bool)

func sseHandler(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ch := make(chan string, 1)
	sseClients[ch] = true

	defer func() {
		delete(sseClients, ch)
		close(ch)
	}()

	for {
		select {
		case msg := <-ch:
			fmt.Fprintf(w, "data: %s\n\n", msg)
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}

func BroadcastReload() {
	for ch := range sseClients {
		ch <- "reload"
	}
}
