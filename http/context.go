package http

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"fbnoi.com/gonet/http/binding"
	"fbnoi.com/gonet/http/render"
	"fbnoi.com/httprouter"
	"fbnoi.com/template"
)

type Context struct {
	context.Context

	Request        *http.Request
	ResponseWriter http.ResponseWriter
	Engine         *Engine
	RouteParams    httprouter.Params

	Error error

	store map[string]any
}

func (ctx *Context) HTML(path string, ps template.Params, code int) {
	writeStatus(ctx.ResponseWriter, code)
	h := render.HTML{ViewPath: path, Params: ps}
	ctx.Error = h.Render(ctx.ResponseWriter)
}

func (ctx *Context) JSON(j *render.JSON, code int) {
	writeStatus(ctx.ResponseWriter, code)
	ctx.Error = j.Render(ctx.ResponseWriter)
}

func (ctx *Context) XML(d any, code int) {
	writeStatus(ctx.ResponseWriter, code)
	x := render.XML{Data: d}
	ctx.Error = x.Render(ctx.ResponseWriter)
}

func (ctx *Context) String(code int, format string, values ...any) {
	writeStatus(ctx.ResponseWriter, code)
	s := render.String{Format: format, Data: values}
	ctx.Error = s.Render(ctx.ResponseWriter)
}

func (ctx *Context) Bytes(code int, contentType string, data ...[]byte) {
	writeStatus(ctx.ResponseWriter, code)
	d := render.Data{
		ContentType: contentType,
		Data:        data,
	}
	ctx.Error = d.Render(ctx.ResponseWriter)
}

func (ctx *Context) Redirect(code int, location string) {
	writeStatus(ctx.ResponseWriter, code)
	r := render.Redirect{
		Code:     code,
		Location: location,
		Request:  ctx.Request,
	}
	ctx.Error = r.Render(ctx.ResponseWriter)
}

func (ctx *Context) RedirectToRoute(code int, name string, ps httprouter.Params) {
	url := ctx.Engine.router.GeneratePath(name, ps)
	ctx.Redirect(code, url)
}

func (ctx *Context) Post(name string) string {
	return ctx.Request.PostForm.Get(name)
}

func (ctx *Context) PostInt(name string) (int, error) {
	i := ctx.Post(name)

	return strconv.Atoi(i)
}

func (ctx *Context) PostBool(name string) (bool, error) {
	i := ctx.Post(name)

	return strconv.ParseBool(i)
}

func (ctx *Context) PostFloat(name string) (float64, error) {
	i := ctx.Post(name)

	return strconv.ParseFloat(i, 64)
}

func (ctx *Context) PostSlice(name, sep string) []string {
	i := ctx.Post(name)

	return strings.Split(i, sep)
}

func (ctx *Context) GetQuery(name string) string {
	return ctx.Request.URL.Query().Get(name)
}

func (ctx *Context) GetInt(name string) (int, error) {
	i := ctx.GetQuery(name)

	return strconv.Atoi(i)
}

func (ctx *Context) GetBool(name string) (bool, error) {
	i := ctx.GetQuery(name)

	return strconv.ParseBool(i)
}

func (ctx *Context) GetFloat(name string) (float64, error) {
	i := ctx.GetQuery(name)

	return strconv.ParseFloat(i, 64)
}

func (ctx *Context) GetSlice(name, sep string) []string {
	i := ctx.GetQuery(name)

	return strings.Split(i, sep)
}

func (ctx *Context) Bind(obj any) error {
	b := binding.Default(ctx.Request.Method, ctx.Request.Header.Get("Content-Type"))

	return ctx.BindWith(obj, b)
}

func (ctx *Context) BindWith(obj any, b binding.BindingInterface) error {
	return b.Bind(ctx.Request, obj)
}

func (ctx *Context) Set(key string, value any) {
	ctx.store[key] = value
}

func (ctx *Context) Get(key string) (val any, ok bool) {
	val, ok = ctx.store[key]

	return
}

func writeStatus(w http.ResponseWriter, code int) {
	w.WriteHeader(code)
}
