package render

import (
	"fmt"
	"io"
	"net/http"

	"github.com/pkg/errors"
)

var plain_content_type = "text/plain; charset=utf-8"

type String struct {
	Format string
	Data   []interface{}
}

func (s String) Render(w http.ResponseWriter) (err error) {
	writeHeader(w, plain_content_type)
	if len(s.Data) > 0 {
		_, err = fmt.Fprintf(w, s.Format, s.Data...)
	} else {
		_, err = io.WriteString(w, s.Format)
	}
	if err != nil {
		err = errors.WithStack(err)
	}
	return
}
