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
const indexFile = "index.html"

func readFS(base string, path string) (fs.File, error) {
	filePath := filepath.Join(base, path)
	f, err := dist.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("file not found: %s", filePath)
	}
	fInfo, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("file not found: %s", filePath)
	}
	if fInfo.IsDir() {
		return nil, fmt.Errorf("file not found: %s", filePath)
	}
	return f, nil
}

func getContentType(path string) string {
	ext := filepath.Ext(path)
	return mime.TypeByExtension(ext)
}

func hostFile(w http.ResponseWriter, r *http.Request, file fs.File) {
	contentType := getContentType(r.URL.Path)
	w.Header().Set("Content-Type", contentType)

	if _, err := io.Copy(w, file); err != nil {
		fmt.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func buildRouter() *chi.Mux {
	router := chi.NewRouter()
	router.NotFound(func(w http.ResponseWriter, r *http.Request) {
		file, err := readFS(assetsBasePath, r.URL.Path)
		if err != nil {
			file, err := readFS(assetsBasePath, indexFile)
			if err != nil {
				http.Error(w, "Not Found", http.StatusNotFound)
				return
			}
			hostFile(w, r, file)
			return
		}
		defer file.Close()

		fmt.Println("hosting file")
		hostFile(w, r, file)
	})
	return router
}
