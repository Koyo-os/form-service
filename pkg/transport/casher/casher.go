package casher

import (
	"context"
	"fmt"

	"github.com/Koyo-os/form-service/pkg/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const FORM_KEY_TEMPLATE = "form:%s"

type Casher struct {
	client *redis.Client
	logger *logger.Logger
}

func Init(client *redis.Client, logger *logger.Logger) *Casher {
	return &Casher{
		client: client,
		logger: logger,
	}
}

func (c *Casher) AddToCash(ctx context.Context, key string, payload []byte) error {
	res := c.client.Set(ctx, fmt.Sprintf("form:%s", key), payload, 0)

	if err := res.Err(); err != nil {
		c.logger.Error("failed to cash payload with",
			zap.String("key", key),
			zap.Error(err),
		)
		return err
	}

	return nil
}

func (c *Casher) GetCashFor(ctx context.Context, key string) ([]byte, error) {
	res := c.client.Get(ctx, fmt.Sprintf(FORM_KEY_TEMPLATE, key))
	if err := res.Err(); err != nil {
		c.logger.Error("error get cash",
			zap.String("key", key),
			zap.Error(err),
		)
		return nil, err
	}

	data, err := res.Bytes()
	if err != nil {
		c.logger.Error("error get cashed bytes",
			zap.String("key", key),
			zap.Error(err),
		)
		return nil, err
	}

	return data, nil
}
