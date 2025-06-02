// Package casher provides Redis-based caching functionality for storing and retrieving form data
package casher

import (
	"context"
	"fmt"

	"github.com/Koyo-os/form-service/pkg/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// FORM_KEY_TEMPLATE defines the format for Redis keys
// It prefixes all form keys with "form:" to create a namespace
const FORM_KEY_TEMPLATE = "form:%s"

// Casher handles caching operations using Redis as the backend
// Note: The name could be "Cacher" for better spelling, but maintaining existing naming
type Casher struct {
	client *redis.Client  // Redis client for storage operations
	logger *logger.Logger // Logger for error tracking and debugging
}

func (c *Casher) RemoveFromCash(ctx context.Context, key string) error {
	res := c.client.Del(ctx, key)

	if res.Err() != nil {
		c.logger.Error("error delete from redis",
			zap.String("key", key),
			zap.Error(res.Err()))
	}

	return nil
}

// Init creates a new Casher instance with the provided Redis client and logger
// This is a simple constructor that doesn't require error handling
func Init(client *redis.Client, logger *logger.Logger) *Casher {
	return &Casher{
		client: client,
		logger: logger,
	}
}

func (c *Casher) Close() error {
	return c.client.Close()
}

func (c *Casher) IsHealthy() bool {
	return c.client.Ping(context.Background()).Err() == nil
}

// AddToCash stores a payload in Redis using the provided key
// The payload is stored with no expiration time (persistence until explicit deletion)
// Parameters:
//   - ctx: Context for cancellation and timeouts
//   - key: Unique identifier for the form data
//   - payload: Raw bytes of the data to be cached
//
// Returns an error if the Redis operation fails
func (c *Casher) AddToCash(ctx context.Context, key string, payload any) error {
	// Format the key using the template and store the payload
	res := c.client.Set(ctx, fmt.Sprintf(FORM_KEY_TEMPLATE, key), payload, 0)

	if err := res.Err(); err != nil {
		c.logger.Error("failed to cash payload with",
			zap.String("key", key),
			zap.Error(err),
		)
		return err
	}

	return nil
}

// GetCashFor retrieves cached data from Redis for the specified key
// Parameters:
//   - ctx: Context for cancellation and timeouts
//   - key: Unique identifier for the form data to retrieve
//
// Returns:
//   - []byte: The cached data if found
//   - error: Error if the retrieval fails or the key doesn't exist
//
// The function handles two potential error cases:
//  1. Redis operation failure
//  2. Byte conversion failure
func (c *Casher) GetCashFor(ctx context.Context, key string) ([]byte, error) {
	// Attempt to retrieve the data from Redis
	res := c.client.Get(ctx, fmt.Sprintf(FORM_KEY_TEMPLATE, key))
	if err := res.Err(); err != nil {
		c.logger.Error("error get cash",
			zap.String("key", key),
			zap.Error(err),
		)
		return nil, err
	}

	// Convert the Redis result to bytes
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
