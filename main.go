package main

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"mime"
	"net/http"
	"path/filepath"

	"github.com/go-chi/chi/v5"
)

// distフォルダをFileSystemとして扱う
//
//go:embed dist
var dist embed.FS

func main() {
	router := buildRouter()

	server := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	if err := server.ListenAndServe(); err != nil {
		panic(err)
	}
}

const assetsBasePath = "dist"

func readFS(base string, path string) (fs.File, error) {
	filePath := filepath.Join(base, path)
	return dist.Open(filePath)
}

func getContentType(path string) string {
	ext := filepath.Ext(path)
	return mime.TypeByExtension(ext)
}

func buildRouter() *chi.Mux {
	router := chi.NewRouter()
	router.NotFound(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r.URL.Path)

		file, err := readFS(assetsBasePath, r.URL.Path)
		if err != nil {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		defer file.Close()

		fInfo, err := file.Stat()
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		if fInfo.IsDir() {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}

		contentType := getContentType(r.URL.Path)
		w.Header().Set("Content-Type", contentType)

		if _, err := io.Copy(w, file); err != nil {
			fmt.Println(err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	})
	return router
}
