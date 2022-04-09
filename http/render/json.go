package render

import (
	"encoding/json"
	"io"

	"github.com/pkg/errors"
)

type JSON map[string]interface{}

func (j JSON) Render(w io.Writer) error {
	return writeJSON(w, j)
}

func writeJSON(w io.Writer, data interface{}) (err error) {
	var jsonBytes []byte
	if jsonBytes, err = json.Marshal(data); err != nil {
		err = errors.WithStack(err)
		return
	}
	if _, err = w.Write(jsonBytes); err != nil {
		err = errors.WithStack(err)
	}
	return
}
