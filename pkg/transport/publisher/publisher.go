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

type Publisher struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	logger  *logger.Logger
	cfg     *config.Config
}

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

func (p *Publisher) Close() error {
	if err := p.channel.Close(); err != nil {
		p.logger.Error("error closing channel", zap.Error(err))
	}
	return p.conn.Close()
}

func (p *Publisher) Publish(poll any, routingKey string) error {
	pollJson, err := json.Marshal(poll)
	if err != nil {
		p.logger.Error("error encode poll for publish", zap.Error(err))
		return err
	}

	event := entity.NewEvent(routingKey, pollJson)

	eventJson, err := json.Marshal(event)
	if err != nil {
		p.logger.Error("error encode event for publish",
			zap.String("event_id", event.ID),
			zap.Error(err),
		)
		return err
	}

	err = p.channel.Publish(
		p.cfg.OutputExcange, // exchange
		routingKey,          // routing key
		false,               // mandatory
		false,               // immediate
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

	p.logger.Info("successfully published event",
		zap.String("event_id", event.ID),
	)

	return nil
}
