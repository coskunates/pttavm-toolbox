package rabbitmq_client

import (
	"context"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"my_toolbox/config"
	"my_toolbox/library/log"
	"net/url"
	"sync"
	"time"
)

type RabbitMQ struct {
	mutex      *sync.RWMutex
	connection *amqp.Connection
}

var (
	rabbitConfig config.RabbitMQConfig
)

// InitWithConfig config ile RabbitMQ client'ı başlatır
func InitWithConfig(cfg config.RabbitMQConfig) {
	rabbitConfig = cfg
}

func NewRabbitMQClient() (*RabbitMQ, error) {
	rabbitMQ := &RabbitMQ{
		mutex: &sync.RWMutex{},
	}

	if err := rabbitMQ.connect(); err != nil {
		log.GetLogger().Error("Feed log rabbitmq connection error : ", err)
		return nil, err
	}

	return rabbitMQ, nil
}

func (rc *RabbitMQ) connect() error {
	rc.mutex.Lock()
	defer rc.mutex.Unlock()

	// Config'den değerleri al
	connectionName := rabbitConfig.ConnectionName
	virtualHost := rabbitConfig.VirtualHost
	username := rabbitConfig.Username
	password := rabbitConfig.Password
	host := rabbitConfig.Host
	port := rabbitConfig.Port

	// Eğer config set edilmemişse panic at
	if host == "" {
		panic("rabbitConfig.Host is required")
	}

	amqpConfig := amqp.Config{
		Properties: amqp.Table{
			"connection_name": connectionName,
		},
		Heartbeat: 30 * time.Second,
		Vhost:     virtualHost,
	}

	connectionUrl := fmt.Sprintf("amqp://%s:%s@%s:%d", username, url.QueryEscape(password), host, port)
	var err error

	rc.connection, err = amqp.DialConfig(connectionUrl, amqpConfig)
	if err != nil {
		_ = rc.Shutdown()
		return err
	}

	log.GetLogger().Info(fmt.Sprintf("RabbitMQ connected: %s@%s:%d/%s", username, host, port, virtualHost))
	go rc.connectionListener()

	return nil
}

func (rc *RabbitMQ) connectionListener() {
	err := <-rc.connection.NotifyClose(make(chan *amqp.Error))
	if err != nil {
		log.GetLogger().Error("connection closed:", err)
	}

	reconnectionInterval := time.Duration(rabbitConfig.ReconnectionInterval) * time.Second
	reconnectionAttempt := rabbitConfig.ReconnectionAttempt

	// Default değerler
	if reconnectionInterval == 0 {
		reconnectionInterval = 10 * time.Second
	}
	if reconnectionAttempt == 0 {
		reconnectionAttempt = 3
	}

	ticker := time.NewTicker(reconnectionInterval)
	defer ticker.Stop()

	var attempt uint
	for range ticker.C {
		attempt++

		log.GetLogger().Info(fmt.Sprintf("reconnection attempt: %d / %d", attempt, reconnectionAttempt))

		if connectionErr := rc.connect(); connectionErr == nil {
			log.GetLogger().Info(fmt.Sprintf("reconnected: %d / %d", attempt, reconnectionAttempt))
			return
		}

		if attempt >= reconnectionAttempt {
			log.GetLogger().Info(fmt.Sprintf("reconnection failed: %d / %d", attempt, reconnectionAttempt))
			return
		}
	}

	log.GetLogger().Error("connection listener explicitly closed the connection", err)
}

func (rc *RabbitMQ) QueueDeclare(queue QueueConfig) error {
	if rc.connection == nil {
		return fmt.Errorf("there is no active connection")
	}

	channel, err := rc.connection.Channel()
	if err != nil {
		log.GetLogger().Error("channel get error", err)
		return err
	}

	err = channel.ExchangeDeclare(queue.Exchange.Name, queue.Exchange.Type, queue.Exchange.Durable, queue.Exchange.AutoDelete, queue.Exchange.Internal, queue.Exchange.NoWait, queue.Exchange.Arguments)
	if err != nil {
		log.GetLogger().Error("exchange declare error", err)
	}

	_, err = channel.QueueDeclare(queue.Name, queue.Durable, queue.AutoDelete, queue.Exclusive, queue.NoWait, queue.Arguments)
	if err != nil {
		log.GetLogger().Error("queue declare error", err)
	}

	err = channel.QueueBind(queue.Name, queue.RoutingKey, queue.Exchange.Name, queue.NoWait, queue.Exchange.Arguments)
	if err != nil {
		log.GetLogger().Error("queue bind error", err)
	}

	return nil
}

func (rc *RabbitMQ) Shutdown() error {
	if rc.connection != nil {
		if err := rc.connection.Close(); err != nil {
			return fmt.Errorf("shutdown: %w", err)
		}
	}

	return nil
}

func (rc *RabbitMQ) Publish(headers amqp.Table, delivery *amqp.Delivery) error {
	rc.mutex.RLock()
	defer rc.mutex.RUnlock()

	if rc.connection == nil {
		return fmt.Errorf("connection is nil")
	}

	chn, err := rc.connection.Channel()
	if err != nil {
		return err
	}

	defer func(chn *amqp.Channel) {
		err = chn.Close()
		if err != nil {
			log.GetLogger().Error("publish error", err)
		}
	}(chn)

	if err = chn.Confirm(false); err != nil {
		return fmt.Errorf("publish on new channel: confirm mode: %w", err)
	}

	err = chn.PublishWithContext(context.Background(), delivery.Exchange, delivery.RoutingKey, false, false, amqp.Publishing{
		Headers:      headers,
		DeliveryMode: amqp.Persistent,
		MessageId:    delivery.MessageId,
		Body:         delivery.Body,
	})
	if err != nil {
		return fmt.Errorf("publish on new channel: confirm mode: %w", err)
	}

	return nil
}
