package render

import (
	"net/http"

	"github.com/pkg/errors"
)

type Data struct {
	ContentType string
	Data        [][]byte
}

func (d Data) Render(w http.ResponseWriter) (err error) {
	writeHeader(w, d.ContentType)
	for _, d := range d.Data {
		if _, err = w.Write(d); err != nil {
			err = errors.WithStack(err)
			return
		}
	}
	return
}
