package render

import (
	"net/http"

	"fbnoi.com/template"
)

const CONTENT_TYPE_HTML = "text/html; charset=utf-8"

type HTML struct {
	ViewPath string
	Params   template.Params
}

func (h HTML) Render(w http.ResponseWriter) (err error) {
	writeHeader(w, CONTENT_TYPE_HTML)
	err = template.Render(h.ViewPath, w, h.Params)

	return
}
