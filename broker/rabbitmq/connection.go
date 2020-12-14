package rabbitmq

import (
	"crypto/tls"
	"github.com/streadway/amqp"
	"regexp"
	"strings"
	"sync"
	"time"
)

var (
	DefaultExchange = Exchange{
		Name: "go-ezio",
	}
	DefaultRabbitURL      = "amqp://guest:guest@localhost:5672"
	DefaultPrefetchCount  = 0
	DefaultPrefetchGlobal = false
	DefaultRequeueOnError = false

	// 自定义dial里不带heartbeat和locale设置
	defaultHeartbeat = 10 * time.Second
	defaultLocale    = "en_US"

	defaultAmqpConfig = amqp.Config{
		Heartbeat: defaultHeartbeat,
		Locale:    defaultLocale,
	}

	dial       = amqp.Dial
	dialTLS    = amqp.DialTLS
	dialConfig = amqp.DialConfig
)

type rabbitMQConn struct {
	Connection      *amqp.Connection
	Channel         *rabbitMQChannel
	ExchangeChannel *rabbitMQChannel
	exchange        Exchange
	url             string
	prefetchCount   int
	prefetchGlobal  bool

	sync.Mutex
	connected bool
	close     chan bool

	waitConnection chan struct{}
}

type Exchange struct {
	// 交换机名称
	Name string
	// 是否持久化
	Durable bool
}

func newRabbitMQConn(ex Exchange, urls []string, prefetchCount int, prefetchGlobal bool) *rabbitMQConn {
	var url string

	if len(urls) > 0 && regexp.MustCompile("^amqp(s)?://.*").MatchString(urls[0]) {
		url = urls[0]
	} else {
		url = DefaultRabbitURL
	}

	ret := &rabbitMQConn{
		exchange:       ex,
		url:            url,
		prefetchCount:  prefetchCount,
		prefetchGlobal: prefetchGlobal,
		close:          make(chan bool),
		waitConnection: make(chan struct{}),
	}
	close(ret.waitConnection)
	return ret
}

func (r *rabbitMQConn) connect(secure bool, config *amqp.Config) error {
	if err := r.tryConnect(secure, config); err != nil {
		return err
	}

	// 已连接
	r.Lock()
	r.connected = true
	r.Unlock()

	go r.reconnect(secure, config)
	return nil
}

func (r *rabbitMQConn) reconnect(secure bool, config *amqp.Config) {
	var connect bool
	// 第一次默认为false

	for {
		if connect {
			if err := r.tryConnect(secure, config); err != nil {
				time.Sleep(1 * time.Second)
				continue
			}

			r.Lock()
			r.connected = true
			r.Unlock()

			close(r.waitConnection)
		}

		connect = true
		notifyClose := make(chan *amqp.Error)
		r.Connection.NotifyClose(notifyClose)

		select {
		case <-notifyClose:
			r.Lock()
			r.connected = false
			r.waitConnection = make(chan struct{})
			r.Unlock()
		case <-r.close:
			return
		}
	}
}

func (r *rabbitMQConn) Connect(secure bool, config *amqp.Config) error {
	r.Lock()

	if r.connected {
		r.Unlock()
		return nil
	}

	select {
	case <-r.close:
		r.close = make(chan bool)
	default:
		// no op
	}

	r.Unlock()

	return r.connect(secure, config)
}

func (r *rabbitMQConn) Close() error {
	r.Lock()
	defer r.Unlock()

	select {
	case <-r.close:
		return nil
	default:
		close(r.close)
		r.connected = false

	}
	return r.Connection.Close()
}

func (r *rabbitMQConn) tryConnect(secure bool, config *amqp.Config) error {
	var err error

	if config == nil {
		config = &defaultAmqpConfig
	}

	url := r.url

	if secure || config.TLSClientConfig != nil || strings.HasPrefix(r.url, "amqps://") {
		if config.TLSClientConfig == nil {
			config.TLSClientConfig = &tls.Config{
				InsecureSkipVerify: true,
			}
		}

		url = strings.Replace(r.url, "amqp://", "amqps://", 1)
	}

	r.Connection, err = dialConfig(url, *config)
	if err != nil {
		return err
	}

	if r.Channel, err = newRabbitMQChannel(r.Connection, r.prefetchCount, r.prefetchGlobal); err != nil {
		return err
	}
	if r.exchange.Durable {
		r.Channel.DeclareDurableExchange(r.exchange.Name)
	} else {
		r.Channel.DeclareExchange(r.exchange.Name)
	}
	r.ExchangeChannel, err = newRabbitMQChannel(r.Connection, r.prefetchCount, r.prefetchGlobal)
	return err
}

func (r *rabbitMQConn) Consume(queue, key string, headers amqp.Table, aArgs amqp.Table, autoAck, durableQueue bool) (*rabbitMQChannel, <-chan amqp.Delivery, error) {
	consumerChannel, err := newRabbitMQChannel(r.Connection, r.prefetchCount, r.prefetchGlobal)
	if err != nil {
		return nil, nil, err
	}

	if durableQueue {
		err = consumerChannel.DeclareQueue(queue, aArgs)
	} else {
		err = consumerChannel.DeclareDurableQueue(queue, aArgs)
	}
	if err != nil {
		return nil, nil, err
	}

	deliveries, err := consumerChannel.ConsumeQueue(queue, autoAck)
	if err != nil {
		return nil, nil, err
	}

	err = consumerChannel.BindQueue(queue, key, r.exchange.Name, headers)
	if err != nil {
		return nil, nil, err
	}
	return consumerChannel, deliveries, nil
}

func (r *rabbitMQConn) Publish(exchange, key string, msg amqp.Publishing) error {
	return r.ExchangeChannel.Publish(exchange, key, msg)
}
