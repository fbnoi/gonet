package response

import (
	"io"
	"log"
	"net/http"
)

const (
	noWritten     = -1
	defaultStatus = http.StatusOK
)

func FromResponse(w http.ResponseWriter) ResponseWriter {
	return &writer{
		ResponseWriter: w,
		size:           noWritten,
		status:         defaultStatus,
	}
}

type ResponseWriter interface {
	http.ResponseWriter

	WriteString(string) (n int, err error)
	WriteHeaderSoft(int)
	WriteContentType(string)
}

type writer struct {
	http.ResponseWriter
	size   int
	status int
	ctype  string
}

func (w *writer) ContentType() string {
	return w.ctype
}

func (w *writer) WriteContentType(ctype string) {
	if ctype != "" {
		header := w.ResponseWriter.Header()
		if header.Get("Content-Type") == "" {
			header.Set("Content-Type", ctype)
			w.ctype = ctype
		}
	}
}

func (w *writer) WriteHeader(code int) {
	if code > 0 && code != w.status {
		if w.size != noWritten {
			log.Printf("[WARNING] Try to override http code %d with %d\n", w.status, code)
		}
		w.status = code
	}
}

func (w *writer) Write(bs []byte) (n int, err error) {
	w.WriteHeaderSoft(defaultStatus)
	n, err = w.ResponseWriter.Write(bs)
	w.size += n
	return
}

func (w *writer) WriteString(str string) (n int, err error) {
	w.WriteHeaderSoft(defaultStatus)
	n, err = io.WriteString(w.ResponseWriter, str)
	w.size += n
	return
}

func (w *writer) WriteHeaderSoft(code int) {
	if w.size == noWritten {
		w.status = code
	}
}
