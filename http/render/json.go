package render

import (
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"
)

const CONTENT_TYPE_JSON = "application/json; charset=utf-8"

type JSON map[string]any

func (j *JSON) Render(w http.ResponseWriter) (err error) {
	writeHeader(w, CONTENT_TYPE_JSON)
	if err = json.NewEncoder(w).Encode(j); err != nil {
		err = errors.WithStack(err)
	}

	return
}
