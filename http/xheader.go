package http

import (
	"net/http"
	"strconv"
	"time"
)

var (
	_header_timeout = "x-timeout"
)

func timeout(r *http.Request) time.Duration {
	t := r.Header.Get(_header_timeout)

	timeout, err := strconv.ParseInt(t, 10, 64)

	if err != nil {
		return _default_timeout
	}

	return time.Duration(timeout * int64(time.Millisecond))
}
