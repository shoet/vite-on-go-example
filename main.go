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

	fmt.Println("Server started at http://localhost:8080")
	if err := server.ListenAndServe(); err != nil {
		panic(err)
	}
}

const assetsBasePath = "dist"
const indexFile = "index.html"

func readFS(path string) (fs.File, error) {
	f, err := dist.Open(path)
	if err != nil {
		return nil, fmt.Errorf("file not found: %s", path)
	}
	fInfo, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("file not found: %s", path)
	}
	if fInfo.IsDir() {
		return nil, fmt.Errorf("file not found: %s", path)
	}
	return f, nil
}

func getContentType(path string) string {
	ext := filepath.Ext(path)
	return mime.TypeByExtension(ext)
}

func hostFile(w http.ResponseWriter, path string, file fs.File) {
	contentType := getContentType(path)
	w.Header().Set("Content-Type", contentType)
	fmt.Println(contentType)

	if _, err := io.Copy(w, file); err != nil {
		fmt.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func buildRouter() *chi.Mux {
	router := chi.NewRouter()
	router.NotFound(func(w http.ResponseWriter, r *http.Request) {
		requestPath := filepath.Join(assetsBasePath, r.URL.Path)
		file, err := readFS(requestPath)
		if err != nil {
			indexPath := filepath.Join(assetsBasePath, indexFile)
			file, err := readFS(indexPath)
			if err != nil {
				fmt.Println(err)
				http.Error(w, "Not Found", http.StatusNotFound)
				return
			}
			hostFile(w, indexPath, file)
			return
		}
		defer file.Close()

		fmt.Println("hosting file")
		hostFile(w, requestPath, file)
	})
	return router
}
