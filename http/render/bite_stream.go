package render

import (
	"io"

	"github.com/pkg/errors"
)

type Data [][]byte

func (d Data) Render(w io.Writer) (err error) {
	for _, d := range d {
		if _, err = w.Write(d); err != nil {
			err = errors.WithStack(err)
			return
		}
	}
	return
}
