package casher // import "github.com/Koyo-os/form-service/pkg/transport/casher"

Package casher provides Redis-based caching functionality for storing and
retrieving form data

CONSTANTS

const FORM_KEY_TEMPLATE = "form:%s"
    FORM_KEY_TEMPLATE defines the format for Redis keys It prefixes all form
    keys with "form:" to create a namespace


TYPES

type Casher struct {
        // Has unexported fields.
}
    Casher handles caching operations using Redis as the backend Note: The name
    could be "Cacher" for better spelling, but maintaining existing naming

func Init(client *redis.Client, logger *logger.Logger) *Casher
    Init creates a new Casher instance with the provided Redis client and logger
    This is a simple constructor that doesn't require error handling

func (c *Casher) AddToCash(ctx context.Context, key string, payload any) error
    AddToCash stores a payload in Redis using the provided key The payload
    is stored with no expiration time (persistence until explicit deletion)
    Parameters:
      - ctx: Context for cancellation and timeouts
      - key: Unique identifier for the form data
      - payload: Raw bytes of the data to be cached

    Returns an error if the Redis operation fails

func (c *Casher) Close() error

func (c *Casher) GetCashFor(ctx context.Context, key string) ([]byte, error)
    GetCashFor retrieves cached data from Redis for the specified key
    Parameters:
      - ctx: Context for cancellation and timeouts
      - key: Unique identifier for the form data to retrieve

    Returns:
      - []byte: The cached data if found
      - error: Error if the retrieval fails or the key doesn't exist

    The function handles two potential error cases:
     1. Redis operation failure
     2. Byte conversion failure

func (c *Casher) IsHealthy() bool

func (c *Casher) RemoveFromCash(ctx context.Context, key string) error