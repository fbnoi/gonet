package render

import (
	"fmt"
	"io"

	"github.com/pkg/errors"
)

type String struct {
	Format string
	Data   []interface{}
}

func (r String) Render(w io.Writer) error {
	return writeString(w, r.Format, r.Data)
}

func writeString(w io.Writer, format string, data []interface{}) (err error) {
	if len(data) > 0 {
		_, err = fmt.Fprintf(w, format, data...)
	} else {
		_, err = io.WriteString(w, format)
	}

	if err != nil {
		err = errors.WithStack(err)
	}
	return
}
