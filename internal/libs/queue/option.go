package queue

import "time"

type OptionFunc func(o *Option)

type Option struct {
	expire    time.Duration
	retention time.Duration
}

func WithExpire(expire time.Duration) OptionFunc {
	return func(o *Option) {
		o.expire = expire
	}
}

func WithRetention(retention time.Duration) OptionFunc {
	return func(o *Option) {
		o.retention = retention
	}
}

func configure(opts ...OptionFunc) *Option {
	o := &Option{
		expire:    0, // 0 means no expiration
		retention: defaultRetention,
	}

	for _, opt := range opts {
		opt(o)
	}

	return o
}
