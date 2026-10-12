package ui

import (
	"embed"
	"io/fs"
)

//go:embed all:dist
var distFS embed.FS

// FS returns a sub-filesystem rooted at dist/
func FS() (fs.FS, error) {
	return fs.Sub(distFS, "dist")
}
