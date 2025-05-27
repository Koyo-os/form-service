// Package consumer provides RabbitMQ consumer functionality for handling message queues
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

// QUEUE_TYPE defines the exchange type for RabbitMQ
// "direct" means messages are routed to queues based on the exact match of routing keys
const QUEUE_TYPE = "direct"

// Consumer represents a RabbitMQ consumer client
// It maintains connection, channel, and configuration details needed for message consumption
type Consumer struct {
	conn      *amqp.Connection // RabbitMQ connection instance
	channel   *amqp.Channel    // Channel for communication with RabbitMQ
	logger    *logger.Logger   // Logger instance for error and info logging
	cfg       *config.Config   // Configuration settings
	exchanges map[string]bool  // Track declared exchanges
}

// Init creates and initializes a new Consumer instance
// Returns an error if the channel creation fails
func Init(cfg *config.Config, logger *logger.Logger, conn *amqp.Connection) (*Consumer, error) {
	channel, err := conn.Channel()
	if err != nil {
		logger.Error("failed to open channel", zap.Error(err))
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

// Subscribe sets up a queue and binds it to an exchange with the specified routing key
// This method handles both queue declaration and queue binding operations
func (p *Consumer) Subscribe(exchange, routingKey, queueName string) error {
	// Declare the queue with specified parameters
	_, err := p.channel.QueueDeclare(
		queueName, // name of the queue
		true,      // durable: queue survives broker restart
		false,     // autoDelete: queue is deleted when last consumer unsubscribes
		false,     // exclusive: queue only accessible by connection that created it
		false,     // noWait: don't wait for server confirmation
		nil,       // args: additional arguments
	)
	if err != nil {
		return err
	}

	// Bind the queue to the exchange using the routing key
	err = p.channel.QueueBind(
		queueName,  // name of the queue to bind
		routingKey, // key used for routing messages
		exchange,   // name of the exchange to bind to
		false,      // noWait: wait for server confirmation
		nil,        // args: additional arguments
	)

	return err
}

func (p *Consumer) Close() error {
	if err := p.channel.Close(); err != nil {
		p.logger.Error("error closing channel", zap.Error(err))
	}

	return p.conn.Close()
}

// ConsumeMessages starts consuming messages from RabbitMQ
// It implements automatic reconnection and message processing in an infinite loop
// Messages are decoded into Events and sent to the provided output channel
func (p *Consumer) ConsumeMessages(outputChan chan entity.Event) {
	for {
		// Check connection status and attempt reconnection if needed
		if p.conn.IsClosed() {
			p.logger.Error("rabbitmq connection is closed, attempting to reconnect...")
			if err := p.reconnect(); err != nil {
				p.logger.Error("failed to reconnect", zap.Error(err))
				time.Sleep(5 * time.Second)
				continue
			}
		}

		// Rebind all exchanges after reconnection
		for exchange := range p.exchanges {
			err := p.channel.QueueBind(
				p.cfg.Queue.Request,
				QUEUE_TYPE,
				exchange,
				false,
				nil,
			)
			if err != nil {
				p.logger.Error("failed to bind queue to exchange",
					zap.String("exchange", exchange),
					zap.Error(err))
			}
		}

		// Start consuming messages
		msgs, err := p.channel.Consume(
			p.cfg.Queue.Request, // queue to consume from
			"",                  // consumer identifier
			true,                // auto-acknowledge messages
			false,               // exclusive consumer
			false,               // no-local flag
			false,               // no-wait flag
			nil,                 // arguments
		)
		if err != nil {
			p.logger.Error("failed to register consumer", zap.Error(err))
			time.Sleep(5 * time.Second)
			continue
		}

		p.logger.Info("successfully connected to RabbitMQ, waiting for messages...")

		// Process incoming messages
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
				zap.Time("timestamp", event.Timestamp))

			outputChan <- event
		}

		p.logger.Warn("rabbitmq channel closed, reconnecting...")
		time.Sleep(5 * time.Second)
	}
}

// reconnect handles the reconnection logic when the RabbitMQ connection is lost
// It re-establishes the connection, recreates the channel, and redeclares all exchanges
func (p *Consumer) reconnect() error {
	var err error

	// Close existing channel if present
	if p.channel != nil {
		p.channel.Close()
	}

	// Establish new connection
	p.conn, err = amqp.Dial(p.cfg.Urls.Rabbitmq)
	if err != nil {
		return err
	}

	// Create new channel
	p.channel, err = p.conn.Channel()
	if err != nil {
		p.conn.Close()
		return err
	}

	// Redeclare all exchanges
	for exchange := range p.exchanges {
		err = p.channel.ExchangeDeclare(
			exchange,   // exchange name
			QUEUE_TYPE, // exchange type
			true,       // durable
			false,      // auto-delete
			false,      // internal
			false,      // no-wait
			nil,        // arguments
		)
		if err != nil {
			p.logger.Error("failed to redeclare exchange",
				zap.String("exchange", exchange),
				zap.Error(err))
		}
	}

	return nil
}
