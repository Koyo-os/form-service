package consumer

import (
	"encoding/json"
	"time"

	"github.com/Koyo-os/form-service/internal/entity"
	"github.com/Koyo-os/form-service/pkg/config"
	"github.com/Koyo-os/form-service/pkg/logger"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

type Consumer struct {
	conn      *amqp.Connection
	channel   *amqp.Channel
	logger    *logger.Logger
	cfg       *config.Config
	exchanges map[string]bool
}

func Init(cfg *config.Config, logger *logger.Logger, conn *amqp.Connection) (*Consumer, error) {
	channel, err := conn.Channel()
	if err != nil {
		logger.Error("failed to open channel", zap.Error(err))
		conn.Close()
		return nil, err
	}

	if err = channel.ExchangeDeclare(
		cfg.RequestExchange,
		"fanout",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		channel.Close()
		conn.Close()
		return nil, err
	}

	return &Consumer{
		conn:      conn,
		channel:   channel,
		logger:    logger,
		cfg:       cfg,
		exchanges: make(map[string]bool),
	}, nil
}

func (p *Consumer) Subscribe(queueName, exchange, routingKey string) error {
	_, err := p.channel.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // autoDelete
		false,     // exclusive
		false,     // noWait
		nil,       // args
	)
	if err != nil {
		return err
	}

	err = p.channel.QueueBind(
		queueName,  // queue name
		routingKey, // routing key
		exchange,   // exchange name
		false,      // noWait
		nil,        // args
	)

	return err
}

func (p *Consumer) ConsumeMessages(outputChan chan entity.Event) {
	for {
		if p.conn.IsClosed() {
			p.logger.Error("rabbitmq connection is closed, attempting to reconnect...")
			if err := p.reconnect(); err != nil {
				p.logger.Error("failed to reconnect", zap.Error(err))
				time.Sleep(5 * time.Second)
				continue
			}
		}

		queue, err := p.channel.QueueDeclare(
			"que", // name
			true,  // durable
			false, // delete when unused
			false, // exclusive
			false, // no-wait
			nil,   // arguments
		)
		if err != nil {
			p.logger.Error("failed to declare a queue", zap.Error(err))
			time.Sleep(5 * time.Second)
			continue
		}

		for exchange := range p.exchanges {
			err = p.channel.QueueBind(
				queue.Name, // queue name
				"",         // routing key
				exchange,   // exchange
				false,      // no-wait
				nil,        // arguments
			)
			if err != nil {
				p.logger.Error("failed to bind queue to exchange",
					zap.String("exchange", exchange),
					zap.Error(err))
			}
		}

		msgs, err := p.channel.Consume(
			queue.Name, // queue
			"",         // consumer
			true,       // auto-ack
			false,      // exclusive
			false,      // no-local
			false,      // no-wait
			nil,        // args
		)
		if err != nil {
			p.logger.Error("failed to register consumer", zap.Error(err))
			time.Sleep(5 * time.Second)
			continue
		}

		p.logger.Info("successfully connected to RabbitMQ, waiting for messages...")

		for msg := range msgs {
			var event entity.Event
			if err := json.Unmarshal(msg.Body, &event); err != nil {
				p.logger.Error("failed to unmarshal event",
					zap.Error(err),
					zap.ByteString("body", msg.Body))
				continue
			}

			p.logger.Debug("received new event",
				zap.String("event_id", event.ID),
				zap.String("routing_key", event.Type),
				zap.Time("timestamp", event.TimeStamp))

			outputChan <- event
		}

		p.logger.Warn("rabbitmq channel closed, reconnecting...")
		time.Sleep(5 * time.Second)
	}
}

func (p *Consumer) reconnect() error {
	var err error

	if p.channel != nil {
		p.channel.Close()
	}

	p.conn, err = amqp.Dial(p.cfg.RabbitmqUrl)
	if err != nil {
		return err
	}

	p.channel, err = p.conn.Channel()
	if err != nil {
		p.conn.Close()
		return err
	}

	for exchange := range p.exchanges {
		err = p.channel.ExchangeDeclare(
			exchange, // name
			"fanout", // type
			true,     // durable
			false,    // auto-deleted
			false,    // internal
			false,    // no-wait
			nil,      // arguments
		)
		if err != nil {
			p.logger.Error("failed to redeclare exchange",
				zap.String("exchange", exchange),
				zap.Error(err))
		}
	}

	return nil
}
