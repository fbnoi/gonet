package kernel

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"net/http"
	"sync/atomic"

	"github.com/pkg/errors"
)

var (
	defaultTimeout         = 5 * time.Second
	defaultReadTimeout     = 2 * time.Second
	defaultWriteTimeout    = 10 * time.Second
	defaultNotFoundHandler = notFoundHandler()
)

func DefaultServer() (k *Kernel) {
	k = &Kernel{
		RouteTree: NewRouteTree(),
		server:    &atomic.Value{},
		conf: &KernelConfig{
			Timeout:           defaultTimeout,
			ReadTimeout:       defaultReadTimeout,
			WriteTimeout:      defaultWriteTimeout,
			RedirectFixedPath: true,
		},
		notFound: defaultNotFoundHandler,
	}
	k.RouteTree.kernel = k
	return
}

type KernelConfig struct {
	Timeout      time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	// wether to redirect to the fixed path first, if on ,
	// http will redirect to the fixed path if path is dirty.
	// for example: /foo/ will direct to /foo
	RedirectFixedPath bool

	CookieConfig *CookieConfig
}

type Kernel struct {
	*RouteTree

	lock sync.RWMutex
	conf *KernelConfig

	server *atomic.Value // store *http.Server

	// handle request when no route was found
	notFound HandlerFunc
}

func (kernel *Kernel) contextFromHttp(r *http.Request, w http.ResponseWriter) *Context {
	return &Context{
		callIdx: -1,
		Request: r,
		Writer:  w,
		Store:   make(map[string]any),
	}
}

func (kernel *Kernel) SetConfig(conf *KernelConfig) error {
	if conf.Timeout <= 0 {
		return errors.New("[error] timeout config must greater than 0")
	}
	kernel.lock.Lock()
	defer kernel.lock.Unlock()
	kernel.conf = conf

	return nil
}

func (kernel *Kernel) GetConfig() *KernelConfig {
	kernel.lock.RLock()
	defer kernel.lock.RUnlock()
	conf := kernel.conf

	return conf
}

func (kernel *Kernel) SetNotFound(fn HandlerFunc) {
	kernel.notFound = fn
}

// Router return a Router
func (kernel *Kernel) GetRouter() IRouter {
	return kernel.RouteTree
}

// Server return stored http server
func (kernel *Kernel) Server() *http.Server {
	s, ok := kernel.server.Load().(*http.Server)
	if ok {
		return s
	}
	return nil
}

// Shutdown the http server without interrupting active connections.
func (kernel *Kernel) Shutdown(ctx context.Context) error {
	server := kernel.Server()
	if server == nil {
		return errors.New("no server running")
	}
	return errors.WithStack(server.Shutdown(ctx))
}

// Ping is used to set the general HTTP ping handler.
// func (kernel *Kernel) Ping(fn HandlerFunc) {
// 	kernel.GET("/monitor/ping", fn)
// }

// Ping is used to set the general HTTP ping handler.
func (kernel *Kernel) NotFound(fn HandlerFunc) {
	kernel.SetNotFoundHandler(fn)
}

// Run attaches the router to a http.Server and starts listening and serving HTTP requests.
// It is a shortcut for http.ListenAndServe(addr, router)
// Note: this method will block the calling goroutine indefinitely unless an error happens.
func (kernel *Kernel) Run(addr string) (err error) {
	defer func() { log.Println(err) }()
	server := &http.Server{
		Addr:         resolveAddr(addr),
		Handler:      kernel.RouteTree,
		ReadTimeout:  time.Duration(kernel.conf.ReadTimeout),
		WriteTimeout: time.Duration(kernel.conf.WriteTimeout),
	}
	kernel.server.Store(server)
	if err = server.ListenAndServe(); err != nil {
		return errors.Wrapf(err, "port: %v", addr)
	}
	return nil
}

// RunTLS attaches the router to a http.Server and starts listening and serving HTTPS (secure) requests.
// It is a shortcut for http.ListenAndServeTLS(addr, certFile, keyFile, router)
// Note: this method will block the calling goroutine indefinitely unless an error happens.
func (kernel *Kernel) RunTLS(port, certFile, keyFile string) (err error) {
	defer func() { log.Println(err) }()
	server := &http.Server{
		Addr:         resolveAddr(port),
		Handler:      kernel.RouteTree,
		ReadTimeout:  time.Duration(kernel.conf.ReadTimeout),
		WriteTimeout: time.Duration(kernel.conf.WriteTimeout),
	}
	kernel.server.Store(server)
	if err = server.ListenAndServeTLS(certFile, keyFile); err != nil {
		err = errors.Wrapf(err, "tls: %s/%s:%s", port, certFile, keyFile)
	}
	return
}

// RunUnix attaches the router to a http.Server and starts listening and serving HTTP requests
// through the specified unix socket (ie. a file).
// Note: this method will block the calling goroutine indefinitely unless an error happens.
func (kernel *Kernel) RunUnix(file string) (err error) {
	os.Remove(file)
	listener, err := net.Listen("unix", file)
	if err != nil {
		err = errors.Wrapf(err, "unix: %s", file)
		return
	}

	defer func() {
		listener.Close()
		log.Println(err)
	}()

	server := &http.Server{
		Handler:      kernel.RouteTree,
		ReadTimeout:  time.Duration(kernel.conf.ReadTimeout),
		WriteTimeout: time.Duration(kernel.conf.WriteTimeout),
	}
	kernel.server.Store(server)
	if err = server.Serve(listener); err != nil {
		err = errors.Wrapf(err, "unix: %s", file)
	}
	return
}

// RunServer will serve and start listening HTTP requests by given server and listener.
// Note: this method will block the calling goroutine indefinitely unless an error happens.
func (kernel *Kernel) RunServer(server *http.Server, l net.Listener) (err error) {
	server.Handler = kernel.RouteTree
	kernel.server.Store(server)
	defer func() { log.Println(err) }()
	if err = server.Serve(l); err != nil {
		err = errors.Wrapf(err, "listen server: %+v/%+v", server, l)
		return
	}
	return
}

func resolveAddr(port string) string {
	return fmt.Sprintf(":%s", strings.Trim(port, ":"))
}

func notFoundHandler() HandlerFunc {
	return func(ctx *Context) {
		ctx.setHttpCode(http.StatusNotFound)
	}
}
