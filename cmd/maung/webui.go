package main

import (
	"io"
	"io/fs"
	"net/http"
)

func serveWebUI() {
	sub, err := fs.Sub(webFS, "WEBUI")
	if err != nil {
		panic(err)
	}
	http.Handle("/public/",
		http.StripPrefix(
			"/public/",
			http.FileServer(http.FS(sub)),
		),
	)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		f, err := sub.Open("index.html")
		if err != nil {
			http.NotFound(w, r)
			return
		}
		defer f.Close()

		w.Header().Set("Content-Type", "text/html")
		_, _ = io.Copy(w, f)
	})
}
