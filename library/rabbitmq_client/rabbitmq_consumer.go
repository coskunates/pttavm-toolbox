package rabbitmq_client

import (
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"my_toolbox/library/log"
	"strconv"
)

type RabbitMQConsumer struct {
	client    *RabbitMQ
	consumers []AbstractConsumer
}

func (rcc *RabbitMQConsumer) SetClient(client *RabbitMQ) {
	rcc.client = client
}

func (rcc *RabbitMQConsumer) RegisterConsumers(consumers []AbstractConsumer) error {
	rcc.consumers = consumers

	if err := rcc.runConsumerWorkers(rcc.consumers); err != nil {
		return fmt.Errorf("register consumers: %w", err)
	}

	return nil
}

func (rcc *RabbitMQConsumer) runConsumerWorkers(consumers []AbstractConsumer) error {
	for _, consumer := range consumers {
		conf := consumer.GetConfig()

		chn, err := rcc.client.connection.Channel()
		if err != nil {
			if err = rcc.client.connect(); err != nil {
				return fmt.Errorf("consumer get channel: %s: %w", conf.Name, err)
			}
		}

		if err := chn.Qos(conf.PrefetchCount, 0, false); err != nil {
			return fmt.Errorf("qos channel: %s: %w", conf.Name, err)
		}

		go rcc.consumerChannelListener(chn, consumer)

		go func(consumer AbstractConsumer) {
			deliveries, err := chn.Consume(
				conf.Name,
				fmt.Sprintf("%s (%d)", conf.Name, conf.WorkerId),
				conf.AutoAck,
				conf.Exclusive,
				conf.NoLocal,
				conf.NoWait,
				conf.Arguments,
			)
			if err != nil {
				log.GetLogger().Error(fmt.Sprintf("consume channel: %s - %d", conf.Name, conf.WorkerId), err)
			}

			log.GetLogger().Info(fmt.Sprintf("run %s - %d", conf.Name, conf.WorkerId))

			for delivery := range deliveries {
				tryCount := uint8(1)
				if _, ok := delivery.Headers["x-try"]; ok {
					val, _ := strconv.ParseInt(delivery.Headers["x-try"].(string), 10, 64)
					tryCount = uint8(val)
				}
				preRunErr := consumer.PreRun(&delivery)
				if preRunErr != nil {
					log.GetLogger().Error(fmt.Sprintf("pre run error %s - %d", conf.Name, conf.WorkerId), preRunErr)
					rcc.retry(consumer, &delivery, tryCount)
					continue
				}
				runErr := consumer.Run()
				if runErr != nil {
					log.GetLogger().Error(fmt.Sprintf("run error %s - %d", conf.Name, conf.WorkerId), runErr)
					rcc.retry(consumer, &delivery, tryCount)
					continue
				}
				postRunErr := consumer.PostRun()
				if postRunErr != nil {
					log.GetLogger().Error(fmt.Sprintf("post run error %s - %d", conf.Name, conf.WorkerId), postRunErr)
					rcc.retry(consumer, &delivery, tryCount)
					continue
				}

				rcc.ack(consumer, &delivery)
			}

			log.GetLogger().Info(fmt.Sprintf("stop %s - %d", conf.Name, conf.WorkerId))
		}(consumer)
	}

	return nil
}

func (rcc *RabbitMQConsumer) consumerChannelListener(chn *amqp.Channel, consumer AbstractConsumer) {
	err := <-chn.NotifyClose(make(chan *amqp.Error))
	if err != nil && err.Code == amqp.ConnectionForced {
		return
	}

	if err != nil {
		log.GetLogger().Error(fmt.Sprintf("consumer channel listener: closed: %s: ", consumer.GetConfig().Name), err)
	} else {
		log.GetLogger().Info(fmt.Sprintf("consumer channel listener: closed: %s: ", consumer.GetConfig().Name))
	}

	if err := rcc.runConsumerWorkers([]AbstractConsumer{consumer}); err != nil {
		log.GetLogger().Error("consumer channel listener ", err)
	}
}

func (rcc *RabbitMQConsumer) retry(consumer AbstractConsumer, delivery *amqp.Delivery, tryCount uint8) {
	rcc.ack(consumer, delivery)
	if consumer.GetConfig().RetryCount > 0 && tryCount < consumer.GetConfig().RetryCount {
		headers := make(amqp.Table)
		headers["x-delay"] = consumer.GetConfig().RetryDelay
		headers["x-try"] = fmt.Sprintf("%d", tryCount+1)

		err := rcc.client.Publish(headers, delivery)
		if err != nil {
			log.GetLogger().Error("rabbitmq publish error", err)
		}
	}
}

func (rcc *RabbitMQConsumer) ack(consumer AbstractConsumer, delivery *amqp.Delivery) {
	if !consumer.GetConfig().AutoAck {
		if err := delivery.Ack(false); err != nil {
			log.GetLogger().Error("consume: ack delivery: %v\n", err)
		}
	}
}
