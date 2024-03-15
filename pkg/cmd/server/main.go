package main

import (
	"log"
	"net/http"
	"os"
	"word-search-in-files/pkg/internal/app"
	"word-search-in-files/pkg/searcher"
)

func main() {
	root := "./examples"
	fs := os.DirFS(root)

	s := &searcher.Searcher{
		FS: fs,
	}
	a := app.NewApp(s)

	s.Init()

	http.HandleFunc("/files/search", a.SearchHandler)

	log.Fatal(http.ListenAndServe("localhost:8080", nil))
}
