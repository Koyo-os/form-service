// Package publisher provides functionality for publishing events to a message broker
package publisher

import (
	"encoding/json"
	"time"

	"github.com/Koyo-os/form-service/internal/entity"
	"github.com/Koyo-os/form-service/pkg/config"
	"github.com/Koyo-os/form-service/pkg/logger"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

// Publisher handles the publication of events to a message broker
type Publisher struct {
	conn    *amqp.Connection // Connection to the message broker
	channel *amqp.Channel    // Channel for publishing messages
	logger  *logger.Logger   // Logger for error tracking and debugging
	cfg     *config.Config   // Configuration settings
}

// Init creates and initializes a new Publisher instance
// Parameters:
//   - cfg: Application configuration
//   - logger: Logger instance for error tracking
//   - conn: Established AMQP connection
//
// Returns:
//   - *Publisher: Initialized publisher instance
//   - error: Any error that occurred during initialization
func Init(cfg *config.Config, logger *logger.Logger, conn *amqp.Connection) (*Publisher, error) {
	channel, err := conn.Channel()
	if err != nil {
		logger.Error("error opening channel", zap.Error(err))
		conn.Close()
		return nil, err
	}
	return &Publisher{
		conn:    conn,
		channel: channel,
		logger:  logger,
		cfg:     cfg,
	}, nil
}

// Close properly closes the publisher's channel and connection
// Returns an error if closing either the channel or connection fails
func (p *Publisher) Close() error {
	if err := p.channel.Close(); err != nil {
		p.logger.Error("error closing channel", zap.Error(err))
	}
	return p.conn.Close()
}

func (p *Publisher) IsHealthy() bool {
	return !p.conn.IsClosed()
}

// Publish sends a message to the message broker
// Parameters:
//   - poll: Data to be published (will be JSON encoded)
//   - routingKey: Routing key for message delivery
//
// Returns:
//   - error: Any error that occurs during publishing
func (p *Publisher) Publish(poll any, routingKey string) error {
	// Convert the poll data to JSON
	pollJson, err := json.Marshal(poll)
	if err != nil {
		p.logger.Error("error encode poll for publish", zap.Error(err))
		return err
	}

	// Create a new event with the JSON payload
	event := entity.NewEvent(routingKey, pollJson)

	// Convert the event to JSON
	eventJson, err := json.Marshal(event)
	if err != nil {
		p.logger.Error("error encode event for publish",
			zap.String("event_id", event.ID),
			zap.Error(err),
		)
		return err
	}

	// Publish the event to the message broker
	err = p.channel.Publish(
		p.cfg.Exchange.Output, // exchange
		routingKey,            // routing key
		false,                 // mandatory
		false,                 // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        eventJson,
			Timestamp:   time.Now(),
		},
	)
	if err != nil {
		p.logger.Error("error publishing event")
		return err
	}

	// Log successful publication
	p.logger.Info("successfully published event",
		zap.String("event_id", event.ID),
	)

	return nil
}
