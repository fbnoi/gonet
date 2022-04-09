package kernel

import (
	"net/http"
	"time"
)

func cookieFromConfig(conf *CookieConfig) *http.Cookie {
	return &http.Cookie{
		Path:       conf.Path,
		Domain:     conf.Domain,
		Expires:    conf.Expires,
		RawExpires: conf.RawExpires,
		MaxAge:     conf.MaxAge,
		Secure:     conf.Secure,
		HttpOnly:   conf.HttpOnly,
		SameSite:   conf.SameSite,
	}
}

type CookieConfig struct {
	Path       string    // optional
	Domain     string    // optional
	Expires    time.Time // optional
	RawExpires string    // for reading cookies only

	MaxAge   int
	Secure   bool
	HttpOnly bool
	SameSite http.SameSite
}

type WithOption func(*http.Cookie)

func WithSameSite(sameSite http.SameSite) WithOption {
	return func(cookie *http.Cookie) {
		cookie.SameSite = sameSite
	}
}

func WithHttpOnly(httpOnly bool) WithOption {
	return func(cookie *http.Cookie) {
		cookie.HttpOnly = httpOnly
	}
}

func WithSecure(secure bool) WithOption {
	return func(cookie *http.Cookie) {
		cookie.Secure = secure
	}
}

func WithMaxAge(maxAge int) WithOption {
	return func(cookie *http.Cookie) {
		cookie.MaxAge = maxAge
	}
}

func WithRawExpires(rawExpires string) WithOption {
	return func(cookie *http.Cookie) {
		cookie.RawExpires = rawExpires
	}
}

func WithExpires(expire time.Time) WithOption {
	return func(cookie *http.Cookie) {
		cookie.Expires = expire
	}
}

func WithPath(path string) WithOption {
	return func(cookie *http.Cookie) {
		cookie.Path = path
	}
}

func WithDomain(domain string) WithOption {
	return func(cookie *http.Cookie) {
		cookie.Domain = domain
	}
}
