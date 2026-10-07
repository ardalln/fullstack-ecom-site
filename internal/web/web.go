// Package web embeds the static frontend so the API server can serve it.
package web

import (
	"embed"
	"io/fs"
)

//go:embed static
var staticFiles embed.FS

// Static returns the frontend files (index.html, app.js, styles.css),
// rooted at the "static" directory.
func Static() (fs.FS, error) {
	return fs.Sub(staticFiles, "static")
}
