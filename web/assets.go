package adminui

import (
	"embed"
	"io/fs"
	"log"
)

//go:embed dist
var distFS embed.FS

func DistFS() fs.FS {
	sub, err := fs.Sub(distFS, "dist")
	if err != nil {
		log.Fatalf("admin UI embedded filesystem corrupted: %v", err)
	}
	return sub
}
