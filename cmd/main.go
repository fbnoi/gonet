package main

import (
	"time"

	"fbnoi.com/gonet/http/kernel"
)

func main() {
	Server := kernel.DefaultServer()
	Server.GET(`/:path(\d+)`, kernel.NewHandler().Final(end1).SetConfig(&kernel.HandlerConfig{
		Timeout: 2 * time.Second,
	}))
	Server.GET(`/setcookie`, kernel.NewHandler().Final(setCookie))
	Server.Run(":8080")
}

func end1(c *kernel.Context) {
	// start := time.Now()
	// <-c.Done()
	c.String(200, "<html><head></head><body><script>alert(123);</script></body></html>")
}

func setCookie(c *kernel.Context) {
	// start := time.Now()
	// <-c.Done()
	c.SetCookie("SESSIONID", "123123123")
	c.String(200, "<html><head></head><body><script>alert(123);</script></body></html>")
}
