package kernel

import (
	"log"
	"sync"
	"time"
)

type MiddlewareFunc func(*Context, HandlerFunc)
type HandlerFunc func(*Context)

func NewHandler(fns ...func()) *Handler {
	return &Handler{}
}

func WithMiddlewares(mds ...MiddlewareFunc) func() {
	return func() {

	}
}

type HandlerConfig struct {
	Timeout time.Duration
}

type Handler struct {
	lock        sync.RWMutex
	conf        *HandlerConfig
	middlewares []MiddlewareFunc
	endpoint    HandlerFunc
}

func (r *Handler) SetConfig(conf *HandlerConfig) *Handler {
	if conf.Timeout <= 0 {
		log.Println("[warning] set handler timeout failed, it will use the kernel's timeout config")
	}
	r.lock.Lock()
	defer r.lock.Unlock()
	r.conf = conf
	return r
}

func (r *Handler) GetConfig() *HandlerConfig {
	r.lock.RLock()
	defer r.lock.RUnlock()
	return r.conf
}

func (r *Handler) Use(md ...MiddlewareFunc) *Handler {
	r.middlewares = append(r.middlewares, md...)
	return r
}

func (r *Handler) GetMiddlewares() []MiddlewareFunc {
	return r.middlewares
}

func (r *Handler) Final(fn HandlerFunc) *Handler {
	if r.endpoint != nil {
		panic("endpoint already set")
	}
	r.endpoint = fn
	return r
}

func (r *Handler) GetEndpoint() HandlerFunc {
	return r.endpoint
}

func (r *Handler) handle(ctx *Context) {
	handlefunc := r.nextHandleFunc(ctx)
	handlefunc(ctx)
}

func (r *Handler) nextHandleFunc(ctx *Context) HandlerFunc {
	nextmd := r.nextMiddleware(ctx)
	return func(_ctx *Context) {
		nextmd(_ctx, r.nextHandleFunc(_ctx))
	}
}

func (r *Handler) nextMiddleware(ctx *Context) MiddlewareFunc {
	ctx.callIdx++
	var md MiddlewareFunc
	if ctx.callIdx >= len(r.middlewares) {
		md = func(c *Context, hf HandlerFunc) {
			r.endpoint(c)
		}
	} else {
		md = r.middlewares[ctx.callIdx]
	}
	return md
}
