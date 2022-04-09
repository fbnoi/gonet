package kernel

import (
	"context"
	"net/http"

	"fbnoi.com/gonet/http/render"
)

type D map[string]any

const (
	CONTENT_TYPE_HTML   = "text/html"
	CONTENT_TYPE_TEXT   = "text/plain"
	CONTENT_TYPE_JSON   = "application/json"
	CONTENT_TYPE_XML    = "application/xml"
	CONTENT_TYPE_STREAM = "application/octet-stream"
)

type Context struct {
	context.Context
	kernel Kernel

	Writer  http.ResponseWriter
	Request *http.Request

	routeParams Params
	Store       map[string]any

	callIdx int
	err     error
}

func (c *Context) GetParam(name string) any {
	if c.routeParams != nil {
		return c.routeParams.ByName(name)
	}
	return nil
}

func (c *Context) GetParams() Params {
	return c.routeParams
}

func (c *Context) Get(name string) (value any, ok bool) {
	value, ok = c.Store[name]
	return
}

func (c *Context) Set(name string, value any) {
	if c.Store == nil {
		c.Store = make(map[string]any)
	}
	c.Store[name] = value
}

func (c *Context) Cookie(name string) *http.Cookie {
	if cookie, err := c.Request.Cookie(name); err == nil {
		return cookie
	}
	return nil
}

func (c *Context) Cookies() []*http.Cookie {
	return c.Request.Cookies()
}

func (c *Context) SetCookie(name, value string) {
	if name != "" && value != "" {
		cookie := cookieFromConfig(c.kernel.GetConfig().CookieConfig)
		cookie.Name, cookie.Value = name, value
		http.SetCookie(c.Writer, cookie)
	}
}

func (c *Context) setContentType(typ string) {
	header := c.Writer.Header()
	if val := header["Content-Type"]; len(val) == 0 {
		header["Content-Type"] = []string{typ}
	}
}

func (c *Context) setHttpCode(code int) {
	c.Writer.WriteHeader(code)
}

func (c *Context) Redirect(code int, location string) {
	http.Redirect(c.Writer, c.Request, location, code)
}

func (c *Context) XML(code int, data D) {
	c.setContentType(CONTENT_TYPE_XML)
	c.render(code, render.XML(data))
}

func (c *Context) Bytes(code int, data ...[]byte) {
	c.setContentType(CONTENT_TYPE_STREAM)
	c.render(code, render.Data(data))
}

func (c *Context) JSON(code int, data D) {
	c.setContentType(CONTENT_TYPE_JSON)
	c.render(code, render.JSON(data))
}

func (c *Context) HTML(code int, html string) {
	c.setContentType(CONTENT_TYPE_HTML)
	c.render(code, render.String{Format: html})
}

func (c *Context) String(code int, format string, data ...any) {
	c.setContentType(CONTENT_TYPE_TEXT)
	c.render(code, render.String{Format: format, Data: data})
}

func (c *Context) render(code int, render render.Render) {
	c.setHttpCode(code)
	if err := render.Render(c.Writer); err != nil {
		c.err = err
	}
}
