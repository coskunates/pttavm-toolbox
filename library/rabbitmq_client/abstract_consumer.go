package rabbitmq_client

import amqp "github.com/rabbitmq/amqp091-go"

type ConsumerConfig struct {
	Name          string
	AutoAck       bool
	Exclusive     bool
	NoLocal       bool
	NoWait        bool
	Arguments     map[string]interface{}
	WorkerId      int8
	PrefetchCount int
	RetryCount    uint8
	RetryDelay    int
}

type IConsumer interface {
	SetConfig(ConsumerConfig)
	GetConfig() ConsumerConfig
	PreRun(delivery *amqp.Delivery) error
	Run() error
	PostRun() error
}

type AbstractConsumer struct {
	IConsumer
	conf ConsumerConfig
}

func (a *AbstractConsumer) SetConfig(config ConsumerConfig) {
	a.conf = config
}

func (a *AbstractConsumer) GetConfig() ConsumerConfig {
	return a.conf
}
