package client

import (
	"net"
	"net/http"
	"time"
)

const (
	defaultMaxIdleConnections = 100
	defaultKeepAliveTimeout   = 600 * time.Second
	defaultRequestTimeout     = 30 * time.Second
)

type Options struct {
	RequestTimeout     time.Duration
	MaxIdleConnections int
	KeepAliveTimeout   time.Duration
}

type Option func(*Options)

func New(opts ...Option) *http.Client {
	options := configure(opts...)

	transport := &http.Transport{
		Dial:                (&net.Dialer{KeepAlive: options.KeepAliveTimeout}).Dial,
		MaxIdleConnsPerHost: options.MaxIdleConnections,
		ForceAttemptHTTP2:   true,
	}

	return &http.Client{
		Transport: transport,
		Timeout:   options.RequestTimeout,
	}
}

func configure(opts ...Option) *Options {
	options := &Options{
		RequestTimeout:     defaultRequestTimeout,
		KeepAliveTimeout:   defaultKeepAliveTimeout,
		MaxIdleConnections: defaultMaxIdleConnections,
	}

	for _, o := range opts {
		o(options)
	}

	return options
}
