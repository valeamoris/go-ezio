package broker

import "github.com/valeamoris/go-ezio/broker/rabbitmq"

type (
	Broker interface {
		Init(...Option) error
		Options() Options
		Address() string
		Connect() error
		Disconnect() error
		Publish(topic string, m *Message, opts ...PublishOption) error
		Subscribe(topic string, h Handler, opts ...SubscribeOption) (Subscriber, error)
		String() string
	}

	Message struct {
		Header map[string]string
		Body   []byte
	}

	Subscriber interface {
		Options() SubscribeOptions
		Topic() string
		Unsubscribe() error
	}

	Event interface {
		Topic() string
		Message() *Message
		Ack() error
		Error() error
	}

	Handler func(Event) error
)

var (
	DefaultBroker = rabbitmq.NewBroker()
)
