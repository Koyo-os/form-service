// Package consumer provides RabbitMQ consumer functionality for handling message queues
package consumer

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/Koyo-os/form-service/internal/entity"
	"github.com/Koyo-os/form-service/pkg/config"
	"github.com/Koyo-os/form-service/pkg/logger"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

const (
	// EXCHANGE_TYPE defines the exchange type for RabbitMQ
	// "direct" means messages are routed to queues based on the exact match of routing keys
	EXCHANGE_TYPE = "direct"
	
	// Default retry settings
	DEFAULT_RECONNECT_DELAY = 5 * time.Second
	DEFAULT_RETRY_ATTEMPTS  = 3
)

// Consumer represents a RabbitMQ consumer client
// It maintains connection, channel, and configuration details needed for message consumption
type Consumer struct {
	conn         *amqp.Connection // RabbitMQ connection instance
	channel      *amqp.Channel    // Channel for communication with RabbitMQ
	logger       *logger.Logger   // Logger instance for error and info logging
	cfg          *config.Config   // Configuration settings
	exchanges    map[string]bool  // Track declared exchanges
	mu           sync.RWMutex     // Mutex for thread-safe operations
	isConnected  bool             // Connection status flag
	reconnecting bool             // Reconnection status flag
}

// Init creates and initializes a new Consumer instance
// Returns an error if the channel creation fails
func Init(cfg *config.Config, logger *logger.Logger, conn *amqp.Connection) (*Consumer, error) {
	if cfg == nil || logger == nil || conn == nil {
		return nil, fmt.Errorf("invalid parameters: cfg, logger, and conn cannot be nil")
	}

	consumer := &Consumer{
		conn:        conn,
		logger:      logger,
		cfg:         cfg,
		exchanges:   make(map[string]bool),
		isConnected: true,
	}

	if err := consumer.initializeChannel(); err != nil {
		return nil, fmt.Errorf("failed to initialize channel: %w", err)
	}

	if err := consumer.declareExchange(cfg.Exchange.Request); err != nil {
		consumer.cleanup()
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	return consumer, nil
}

// initializeChannel creates a new channel and sets up basic configuration
func (c *Consumer) initializeChannel() error {
	channel, err := c.conn.Channel()
	if err != nil {
		c.logger.Error("failed to open channel", zap.Error(err))
		return err
	}

	c.channel = channel
	return nil
}

// declareExchange declares an exchange and tracks it
func (c *Consumer) declareExchange(exchangeName string) error {
	if err := c.channel.ExchangeDeclare(
		exchangeName,
		EXCHANGE_TYPE,
		true,  // durable
		false, // auto-delete
		false, // internal
		false, // no-wait
		nil,   // arguments
	); err != nil {
		c.logger.Error("failed to declare exchange", 
			zap.String("exchange", exchangeName), 
			zap.Error(err))
		return err
	}

	c.mu.Lock()
	c.exchanges[exchangeName] = true
	c.mu.Unlock()

	return nil
}

// Subscribe sets up a queue and binds it to an exchange with the specified routing key
// This method handles both queue declaration and queue binding operations
func (c *Consumer) Subscribe(exchange, routingKey, queueName string) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.isConnected {
		return fmt.Errorf("consumer is not connected")
	}

	// Declare the queue with specified parameters
	if _, err := c.channel.QueueDeclare(
		queueName, // name of the queue
		true,      // durable: queue survives broker restart
		false,     // autoDelete: queue is deleted when last consumer unsubscribes
		false,     // exclusive: queue only accessible by connection that created it
		false,     // noWait: don't wait for server confirmation
		nil,       // args: additional arguments
	); err != nil {
		c.logger.Error("failed to declare queue", 
			zap.String("queue", queueName), 
			zap.Error(err))
		return fmt.Errorf("failed to declare queue %s: %w", queueName, err)
	}

	// Bind the queue to the exchange using the routing key
	if err := c.channel.QueueBind(
		queueName,  // name of the queue to bind
		routingKey, // key used for routing messages
		exchange,   // name of the exchange to bind to
		false,      // noWait: wait for server confirmation
		nil,        // args: additional arguments
	); err != nil {
		c.logger.Error("failed to bind queue to exchange", 
			zap.String("queue", queueName),
			zap.String("exchange", exchange),
			zap.String("routing_key", routingKey),
			zap.Error(err))
		return fmt.Errorf("failed to bind queue %s to exchange %s: %w", queueName, exchange, err)
	}

	// Track the exchange
	c.mu.Lock()
	c.exchanges[exchange] = true
	c.mu.Unlock()

	return nil
}

// Close gracefully closes the consumer connection and channel
func (c *Consumer) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.isConnected = false

	var errors []error

	if c.channel != nil {
		if err := c.channel.Close(); err != nil {
			c.logger.Error("error closing channel", zap.Error(err))
			errors = append(errors, fmt.Errorf("channel close error: %w", err))
		}
	}

	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			c.logger.Error("error closing connection", zap.Error(err))
			errors = append(errors, fmt.Errorf("connection close error: %w", err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors during close: %v", errors)
	}

	return nil
}

// IsHealthy checks if the consumer connection is healthy
func (c *Consumer) IsHealthy() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	return c.isConnected && c.conn != nil && !c.conn.IsClosed()
}

// ConsumeMessages starts consuming messages from RabbitMQ
// It implements automatic reconnection and message processing in an infinite loop
// Messages are decoded into Events and sent to the provided output channel
func (c *Consumer) ConsumeMessages(outputChan chan entity.Event) {
	if outputChan == nil {
		c.logger.Error("output channel cannot be nil")
		return
	}

	for {
		if !c.IsHealthy() {
			c.logger.Warn("connection is unhealthy, attempting to reconnect...")
			if err := c.handleReconnection(); err != nil {
				c.logger.Error("failed to reconnect", zap.Error(err))
				time.Sleep(DEFAULT_RECONNECT_DELAY)
				continue
			}
		}

		if err := c.rebindExchanges(); err != nil {
			c.logger.Error("failed to rebind exchanges", zap.Error(err))
			time.Sleep(DEFAULT_RECONNECT_DELAY)
			continue
		}

		if err := c.startConsuming(outputChan); err != nil {
			c.logger.Error("consuming stopped with error", zap.Error(err))
			time.Sleep(DEFAULT_RECONNECT_DELAY)
		}
	}
}

// handleReconnection manages the reconnection process with proper synchronization
func (c *Consumer) handleReconnection() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.reconnecting {
		return fmt.Errorf("reconnection already in progress")
	}

	c.reconnecting = true
	defer func() { c.reconnecting = false }()

	return c.reconnect()
}

// startConsuming handles the actual message consumption
func (c *Consumer) startConsuming(outputChan chan entity.Event) error {
	msgs, err := c.channel.Consume(
		c.cfg.Queue.Request, // queue to consume from
		"",                  // consumer identifier
		true,                // auto-acknowledge messages
		false,               // exclusive consumer
		false,               // no-local flag
		false,               // no-wait flag
		nil,                 // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	c.logger.Info("successfully connected to RabbitMQ, waiting for messages...")

	// Process incoming messages
	for msg := range msgs {
		if err := c.processMessage(msg, outputChan); err != nil {
			c.logger.Error("failed to process message", zap.Error(err))
			// Continue processing other messages even if one fails
		}
	}

	return fmt.Errorf("message channel closed")
}

// processMessage handles individual message processing
func (c *Consumer) processMessage(msg amqp.Delivery, outputChan chan entity.Event) error {
	event := new(entity.Event)
	if err := json.Unmarshal(msg.Body, event); err != nil {
		c.logger.Error("failed to unmarshal event",
			zap.Error(err),
			zap.ByteString("body", msg.Body))
		return fmt.Errorf("failed to unmarshal message: %w", err)
	}

	c.logger.Debug("received new event",
		zap.String("event_id", event.ID),
		zap.String("routing_key", event.Type),
		zap.Time("timestamp", event.Timestamp))

	// Non-blocking send to output channel
	select {
	case outputChan <- *event:
		return nil
	default:
		c.logger.Warn("output channel is full, dropping message",
			zap.String("event_id", event.ID))
		return fmt.Errorf("output channel is full")
	}
}

// rebindExchanges rebinds all tracked exchanges after reconnection
func (c *Consumer) rebindExchanges() error {
	c.mu.RLock()
	exchanges := make([]string, 0, len(c.exchanges))
	for exchange := range c.exchanges {
		exchanges = append(exchanges, exchange)
	}
	c.mu.RUnlock()

	for _, exchange := range exchanges {
		if err := c.channel.QueueBind(
			c.cfg.Queue.Request,
			EXCHANGE_TYPE,
			exchange,
			false,
			nil,
		); err != nil {
			c.logger.Error("failed to bind queue to exchange",
				zap.String("exchange", exchange),
				zap.Error(err))
			return fmt.Errorf("failed to bind exchange %s: %w", exchange, err)
		}
	}

	return nil
}

// reconnect handles the reconnection logic when the RabbitMQ connection is lost
// It re-establishes the connection, recreates the channel, and redeclares all exchanges
func (c *Consumer) reconnect() error {
	c.cleanup()

	// Establish new connection
	conn, err := amqp.Dial(c.cfg.Urls.Rabbitmq)
	if err != nil {
		return fmt.Errorf("failed to dial RabbitMQ: %w", err)
	}

	c.conn = conn

	// Create new channel
	if err := c.initializeChannel(); err != nil {
		c.conn.Close()
		return err
	}

	// Redeclare all exchanges
	c.mu.RLock()
	exchanges := make([]string, 0, len(c.exchanges))
	for exchange := range c.exchanges {
		exchanges = append(exchanges, exchange)
	}
	c.mu.RUnlock()

	for _, exchange := range exchanges {
		if err := c.declareExchange(exchange); err != nil {
			c.cleanup()
			return fmt.Errorf("failed to redeclare exchange %s: %w", exchange, err)
		}
	}

	c.isConnected = true
	c.logger.Info("successfully reconnected to RabbitMQ")
	return nil
}

// cleanup closes existing connections and channels
func (c *Consumer) cleanup() {
	c.isConnected = false

	if c.channel != nil {
		c.channel.Close()
		c.channel = nil
	}

	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
}
