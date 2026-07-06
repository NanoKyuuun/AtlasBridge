package server

import (
	"io/fs"
	"net/http"
	"path"
	"strings"

	adminui "github.com/atlasbridge/atlasbridge/web"
)

var adminDistFS = adminui.DistFS()

func adminUIHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		relPath := strings.TrimPrefix(r.URL.Path, "/admin")
		relPath = strings.TrimPrefix(relPath, "/")

		if relPath == "" || relPath == "index.html" || path.Ext(relPath) == "" {
			serveAdminIndex(w)
			return
		}

		data, err := fs.ReadFile(adminDistFS, relPath)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Content-Type", mimeContentType(relPath))
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		w.Write(data)
	}
}

func serveAdminIndex(w http.ResponseWriter) {
	data, err := fs.ReadFile(adminDistFS, "index.html")
	if err != nil {
		http.Error(w, "admin UI build not found", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Write(data)
}
