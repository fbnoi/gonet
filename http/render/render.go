package render

import "io"

type Render interface {
	Render(io.Writer) error
}
