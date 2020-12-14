package broker

import (
	"context"
	"crypto/tls"
)

type (
	Options struct {
		Addrs  []string
		Secure bool

		// 错误处理函数
		ErrorHandler Handler

		TLSConfig *tls.Config

		Context context.Context
	}

	PublishOptions struct {
		Context context.Context
	}

	SubscribeOptions struct {
		AutoAck bool
		Queue   string
		Context context.Context
	}

	Option func(*Options)

	PublishOption func(*PublishOptions)

	SubscribeOption func(*SubscribeOptions)
)

func PublishContext(ctx context.Context) PublishOption {
	return func(o *PublishOptions) {
		o.Context = ctx
	}
}

func NewSubscribeOptions(opts ...SubscribeOption) SubscribeOptions {
	opt := SubscribeOptions{
		AutoAck: true,
	}

	for _, o := range opts {
		o(&opt)
	}

	return opt
}

// 设置broker地址
func Addrs(addrs ...string) Option {
	return func(o *Options) {
		o.Addrs = addrs
	}
}

// 关闭自动确认
func DisableAutoAck() SubscribeOption {
	return func(o *SubscribeOptions) {
		o.AutoAck = false
	}
}

// 错误处理函数
func ErrorHandler(h Handler) Option {
	return func(o *Options) {
		o.ErrorHandler = h
	}
}

// message共享的queue名称
func Queue(name string) SubscribeOption {
	return func(o *SubscribeOptions) {
		o.Queue = name
	}
}

func Secure(b bool) Option {
	return func(o *Options) {
		o.Secure = b
	}
}

func TLSConfig(t *tls.Config) Option {
	return func(o *Options) {
		o.TLSConfig = t
	}
}

func SubscribeContext(ctx context.Context) SubscribeOption {
	return func(o *SubscribeOptions) {
		o.Context = ctx
	}
}
