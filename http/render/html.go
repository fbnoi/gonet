package render

import (
	"net/http"

	"fbnoi.com/template"
)

const CONTENT_TYPE_HTML = "text/html; charset=utf-8"

type HTML struct {
	viewPath string
	Params   template.Params
}

func (h HTML) Render(w http.ResponseWriter) (err error) {
	writeHeader(w, CONTENT_TYPE_HTML)
	err = template.Render(h.viewPath, w, h.Params)

	return
}
