package render

import (
	"encoding/xml"
	"net/http"

	"github.com/pkg/errors"
)

const content_type_xml = "application/xml; charset=utf-8"

type XML struct {
	Data any
}

func (x XML) Render(w http.ResponseWriter) (err error) {
	writeHeader(w, content_type_xml)
	if err = xml.NewEncoder(w).Encode(x.Data); err != nil {
		err = errors.WithStack(err)
	}
	return
}
