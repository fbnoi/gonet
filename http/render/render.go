package render

import "net/http"

var (
	Json = &JSON{}
)

type Render interface {
	Render(w http.ResponseWriter) error
}

func writeHeader(w http.ResponseWriter, contentType string) {
	if w.Header().Get("Content-Type") == "" {
		w.Header().Set("Content-Type", contentType)
	}
}
