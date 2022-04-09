package kernel

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"

	"fbnoi.com/goutil/collection"
)

// fixme: move this to kernel
var default_not_found_handler = notFoundMethod

func NewRouteTree() *RouteTree {
	return &RouteTree{
		metadata:        make(map[string]map[string]*Handler),
		once:            &sync.Once{},
		notFoundHandler: default_not_found_handler,
		RouteNode: &RouteNode{
			children: &collection.LinkedList[*RouteNode]{},
			root:     true,
			leaf:     false,
		},
	}
}

const (
	METHOD_GET     = http.MethodGet
	METHOD_POST    = http.MethodPost
	METHOD_HEAD    = http.MethodHead
	METHOD_PUT     = http.MethodPut
	METHOD_PATCH   = http.MethodPatch
	METHOD_DELETE  = http.MethodDelete
	METHOD_CONNECT = http.MethodConnect
	METHOD_OPTIONS = http.MethodOptions
	METHOD_TRACE   = http.MethodTrace
)

var allowed_methods = []string{
	METHOD_GET, METHOD_POST, METHOD_HEAD,
	METHOD_PUT, METHOD_PATCH, METHOD_DELETE,
}

// IRouter http router interface.
type IRouter interface {
	Use(...MiddlewareFunc) IRouter

	Group(string, func(*RouteTree), ...MiddlewareFunc)

	Handle(string, string, *Handler) IRouter
	POST(string, *Handler) IRouter
	GET(string, *Handler) IRouter
	HEAD(string, *Handler) IRouter
	PUT(string, *Handler) IRouter
	PATCH(string, *Handler) IRouter
	DELETE(string, *Handler) IRouter
}

// RouteTree keeps a nodetree where stores the registered path and it's
// handler.
// TODO: add name to route as to generate a url by name
type RouteTree struct {
	*RouteNode
	metadata        map[string]map[string]*Handler
	once            *sync.Once
	baseMds         []MiddlewareFunc
	notFoundHandler HandlerFunc
	paramsPool      sync.Pool
	kernel          *Kernel
}

// ServeHTTP serve the http request
func (rt *RouteTree) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	path := r.URL.Path
	fixedPath := cleanPath(path)
	r.URL.Path = fixedPath

	ctx := rt.kernel.contextFromHttp(r, w)

	conf := rt.kernel.GetConfig()

	// if path is dirty, redirect
	if path != fixedPath && conf.RedirectFixedPath {
		ctx.Redirect(http.StatusTemporaryRedirect, r.URL.String())

		return
	}

	// if path endup with '/', redirect
	if len(fixedPath) > 1 && strings.HasSuffix(fixedPath, "/") {
		r.URL.Path = strings.TrimRight(fixedPath, "/")
		ctx.Redirect(http.StatusTemporaryRedirect, r.URL.String())

		return
	}
	node, params := rt.kernel.RouteTree.lookUp(path)
	defer rt.putParams(&params)

	// not found
	if node == nil {
		rt.kernel.notFound(ctx)
		return
	}

	handler, ok := node.getHandlers(r.Method)
	if !ok {
		rt.kernel.notFound(ctx)
		return
	}

	ctx.routeParams = params

	tm := conf.Timeout
	hcon := handler.GetConfig()
	if hcon != nil && hcon.Timeout < tm {
		tm = hcon.Timeout
	}

	var cancel func()
	ctx.Context, cancel = context.WithTimeout(context.Background(), tm)
	defer cancel()

	handler.handle(ctx)
}

func (rt *RouteTree) lookUp(path string) (*RouteNode, Params) {
	node, ps := rt.search(path, rt.getParams)
	if node == nil {
		return nil, nil
	}
	if ps == nil {
		return node, nil
	}
	return node, *ps
}

func (rt *RouteTree) Group(path string, fn func(*RouteTree), mds ...MiddlewareFunc) {
	node := rt.RouteNode.pave(path)
	tree := &RouteTree{
		RouteNode: node,
		metadata:  make(map[string]map[string]*Handler),
		once:      &sync.Once{},
		baseMds:   append(rt.baseMds, mds...),
	}
	fn(tree)
	for method, meta := range tree.metadata {
		if rt.metadata[method] == nil {
			rt.metadata[method] = make(map[string]*Handler)
		}
		for path, handler := range meta {
			rt.metadata[method][path] = handler
		}
	}
}

// Use add middlewares to globle middlewares
// middleware added by Use() won't infect the former added path
func (rt *RouteTree) Use(mds ...MiddlewareFunc) IRouter {
	rt.baseMds = append(rt.baseMds, mds...)
	return rt
}

func (rt *RouteTree) POST(path string, handler *Handler) IRouter {
	return rt.addPath(METHOD_POST, path, handler)
}

func (rt *RouteTree) GET(path string, handler *Handler) IRouter {
	return rt.addPath(METHOD_GET, path, handler)
}

func (rt *RouteTree) HEAD(path string, handler *Handler) IRouter {
	return rt.addPath(METHOD_HEAD, path, handler)
}

func (rt *RouteTree) PUT(path string, handler *Handler) IRouter {
	return rt.addPath(METHOD_HEAD, path, handler)
}

func (rt *RouteTree) PATCH(path string, handler *Handler) IRouter {
	return rt.addPath(METHOD_PATCH, path, handler)
}

func (rt *RouteTree) DELETE(path string, handler *Handler) IRouter {
	return rt.addPath(METHOD_DELETE, path, handler)
}

func (rt *RouteTree) Handle(method, path string, handler *Handler) IRouter {
	return rt.addPath(method, path, handler)
}

func (rt *RouteTree) SetNotFoundHandler(fn HandlerFunc) {
	rt.notFoundHandler = fn
}

func (rt *RouteTree) addPath(method, path string, handler *Handler) *RouteTree {
	if !allowMethod(method) {
		methodNotAllowed(method)
	}
	rt.init()
	if len(rt.baseMds) != 0 {
		handler.middlewares = append(rt.baseMds, handler.middlewares...)
	}
	node := rt.RouteNode.addPath(method, path, handler)
	if rt.metadata[method] == nil {
		rt.metadata[method] = make(map[string]*Handler)
	}
	rt.metadata[method][node.fullPath] = handler
	return rt
}

func (rt *RouteTree) init() {
	rt.once.Do(func() {
		rt.metadata = make(map[string]map[string]*Handler)
		rt.paramsPool.New = func() interface{} {
			ps := make(Params, 0)
			return &ps
		}
	})
}

func (rt *RouteTree) getParams() *Params {
	ps, _ := rt.paramsPool.Get().(*Params)
	*ps = (*ps)[0:0] // reset slice
	return ps
}

func (rt *RouteTree) putParams(ps *Params) {
	if ps != nil {
		rt.paramsPool.Put(ps)
	}
}

func allowMethod(method string) bool {
	for _, m := range allowed_methods {
		if method == m {
			return true
		}
	}
	return false
}

func methodNotAllowed(method string) {
	panic(fmt.Sprintf("method [%s] not allowed", method))
}

func notFoundMethod(ctx *Context) {
	log.Println("404 not fount")
}
