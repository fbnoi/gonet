package render

import (
	"encoding/xml"
	"io"

	"github.com/pkg/errors"
)

type XML map[string]interface{}

func (x XML) Render(w io.Writer) (err error) {
	if err = xml.NewEncoder(w).Encode(x); err != nil {
		err = errors.WithStack(err)
	}
	return
}
