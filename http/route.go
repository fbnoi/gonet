package http

import (
	"context"
	"net/http"
	"strings"

	"fbnoi.com/handler"
	"fbnoi.com/httprouter"
)

func (e *Engine) GET(name, path string, fn func(*Context), mds ...func(*Context, func(*Context))) *Engine {
	return e.Handle(name, "GET", path, fn, mds...)
}

func (e *Engine) POST(name, path string, fn func(*Context), mds ...func(*Context, func(*Context))) *Engine {
	return e.Handle(name, "POST", path, fn, mds...)
}

func (e *Engine) HEAD(name, path string, fn func(*Context), mds ...func(*Context, func(*Context))) *Engine {
	return e.Handle(name, "HEAD", path, fn, mds...)
}

func (e *Engine) PUT(name, path string, fn func(*Context), mds ...func(*Context, func(*Context))) *Engine {
	return e.Handle(name, "PUT", path, fn, mds...)
}

func (e *Engine) PATCH(name, path string, fn func(*Context), mds ...func(*Context, func(*Context))) *Engine {
	return e.Handle(name, "PATCH", path, fn, mds...)
}

func (e *Engine) DELETE(name, path string, fn func(*Context), mds ...func(*Context, func(*Context))) *Engine {
	return e.Handle(name, "DELETE", path, fn, mds...)
}

func (e *Engine) All(name, path string, fn func(*Context), mds ...func(*Context, func(*Context))) *Engine {
	h := wrapHandler(fn, mds...)
	e.router.All(name, path, func(r *http.Request, w http.ResponseWriter, ps httprouter.Params) {
		e.handle(r, w, ps, h)
	})

	return e
}

func (e *Engine) Handle(name, method, path string, fn func(*Context), mds ...func(*Context, func(*Context))) *Engine {
	h := wrapHandler(fn, mds...)
	e.router.Handle(name, method, path, func(r *http.Request, w http.ResponseWriter, ps httprouter.Params) {
		e.handle(r, w, ps, h)
	})

	return e
}

func wrapHandler(fn func(*Context), mds ...func(*Context, func(*Context))) *handler.Handler[*Context] {
	return handler.New[*Context]().Then(mds...).Final(fn)
}

func (e *Engine) handle(r *http.Request, w http.ResponseWriter, ps httprouter.Params, h *handler.Handler[*Context]) {
	conf := e.config()
	rConf, ok := e.routeConfig(ps.GetRoute().RouteName())

	mem, t := conf.MaxMemory, conf.TimeOut
	if ok {
		mem, t = rConf.MaxMemory, rConf.TimeOut
	}
	contentType := r.Header.Get("Content-Type")
	if strings.Contains(contentType, "multipart/form-data") {
		r.ParseMultipartForm(mem)
	} else {
		r.ParseForm()
	}
	var cancel func()
	ctx := &Context{
		Request:        r,
		ResponseWriter: w,
		Engine:         e,
		RouteParams:    ps,
	}
	if t > 0 {
		ctx.Context, cancel = context.WithTimeout(context.Background(), t)
	} else {
		ctx.Context, cancel = context.WithCancel(context.Background())
	}
	defer cancel()

	h.Handle(ctx)
}
