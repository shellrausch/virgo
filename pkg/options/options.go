package options

import "net/http"

// Options contains fields to control a HTTP request.
type Options struct {
	Method          string
	Header          map[string]string
	Body            []byte
	UserAgent       string
	Cookie          string
	TimeoutMs       int
	Concurrency     int
	FollowRedirects bool
}

// New creates and initializes a Options struct.
func New() *Options {
	return &Options{
		Method:          http.MethodGet,
		Header:          map[string]string{},
		Body:            []byte{},
		UserAgent:       "",
		Cookie:          "",
		TimeoutMs:       60000,
		Concurrency:     8,
		FollowRedirects: false,
	}
}
