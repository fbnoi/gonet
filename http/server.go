package http

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"fbnoi.com/httprouter"
	"github.com/pkg/errors"
)

var (
	_default_memory  int64         = 2 << 20
	_default_timeout time.Duration = 1 * time.Second
)

func DefaultEngine() *Engine {
	return &Engine{
		router: httprouter.NewRouteTree(&httprouter.Config{RedirectFixedPath: true}),
		conf:   &Config{_default_memory, _default_timeout},
	}
}

type Config struct {
	MaxMemory int64
	TimeOut   time.Duration
}

type Engine struct {
	server atomic.Value

	router *httprouter.RouteTree

	lock sync.RWMutex
	conf *Config

	rLock        sync.RWMutex
	routeConfigs map[string]*Config
}

func (e *Engine) SetConfig(conf *Config) error {
	if conf.TimeOut < 0 {
		return errors.New("Timeout cannot less than 0.")
	}

	e.lock.Lock()
	defer e.lock.Unlock()
	e.conf = conf

	return nil
}

func (e *Engine) config() (c *Config) {
	e.lock.Lock()
	defer e.lock.Unlock()
	c = e.conf

	return
}

func (e *Engine) SetRouteConfig(name string, conf *Config) error {
	if conf.TimeOut < 0 {
		return errors.New("Timeout cannot less than 0.")
	}

	e.rLock.Lock()
	defer e.rLock.Unlock()
	e.routeConfigs[name] = conf

	return nil
}

func (e *Engine) routeConfig(name string) (c *Config, ok bool) {
	e.rLock.Lock()
	defer e.rLock.Unlock()

	c, ok = e.routeConfigs[name]
	return
}

func (e *Engine) Server() *http.Server {
	s, ok := e.server.Load().(*http.Server)
	if !ok {
		return nil
	}
	return s
}

func (engine *Engine) Shutdown(ctx context.Context) error {
	server := engine.Server()
	if server == nil {
		return errors.New("no server")
	}
	return errors.WithStack(server.Shutdown(ctx))
}

func (e *Engine) Run(port string) (err error) {
	defer func() { log.Println(err) }()
	server := &http.Server{
		Addr:    resolveAddr(port),
		Handler: e.router,
	}
	e.server.Store(server)

	if err = server.ListenAndServe(); err != nil {
		return errors.Wrapf(err, "port: %v", port)
	}

	return
}

func (e *Engine) RunTLS(port, certFile, keyFile string) (err error) {
	defer func() { log.Println(err) }()

	server := &http.Server{
		Addr:    resolveAddr(port),
		Handler: e.router,
	}
	e.server.Store(server)
	if err = server.ListenAndServeTLS(certFile, keyFile); err != nil {
		err = errors.Wrapf(err, "tls: %s/%s:%s", port, certFile, keyFile)
	}

	return
}

func resolveAddr(port string) string {
	return fmt.Sprintf(":%s", strings.Trim(port, ":"))
}
