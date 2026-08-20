package webfs

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed web
var webFS embed.FS

// Handler serves the embedded frontend at /.
func Handler() http.Handler {
	sub, err := fs.Sub(webFS, "web")
	if err != nil {
		// Should never happen for a static embed; fail closed.
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "frontend unavailable", http.StatusInternalServerError)
		})
	}
	return http.FileServer(http.FS(sub))
}

// Files returns the list of embedded web files (for smoke-test verification).
func Files() []string {
	entries, err := fs.ReadDir(webFS, "web")
	if err != nil {
		return nil
	}
	var out []string
	for _, e := range entries {
		out = append(out, e.Name())
	}
	return out
}
