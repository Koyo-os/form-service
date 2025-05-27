package retrier

// RetrierOpts contains configuration options for retry operations.
// This struct is used to specify how many times an operation should be retried
// and the interval between retry attempts.
type RetrierOpts struct {
	Count    uint // Number of retry attempts (0 means no retries)
	Interval uint // Delay between retries in seconds
}

// MultiConnects establishes multiple connections of type T with optional retry logic.
//
// This function creates multiple connections using the provided connection function,
// with the ability to apply retry behavior for each connection attempt when RetrierOpts
// is specified. It's particularly useful for establishing pools of connections where
// individual connections might need retry logic.
//
// Type Parameters:
//   - T: The type of connection being established
//
// Parameters:
//   - count: The number of connections to establish (must be a uint8)
//   - connFunc: The function that creates a single connection (returns T and error)
//   - retrierOpts: Optional retry configuration (nil means no retries)
//
// Returns:
//   - []T: Slice of successfully established connections
//   - error: The first error encountered during connection attempts, if any
//
// Behavior:
//   - If retrierOpts is nil, connections are attempted once without retries
//   - If retrierOpts is provided, each connection attempt will be retried up to Count times
//   - The function fails fast - returns immediately on first connection error
//   - All established connections are returned on success
//
// Example Usage:
//
//	connections, err := MultiConnects(3, dialDatabase, &RetrierOpts{Count: 3, Interval: 1})
func MultiConnects[T any](count uint8, connFunc func() (T, error), retrierOpts *RetrierOpts) ([]T, error) {
	// Initialize slice to hold all connections
	conns := make([]T, count)

	var err error

	// Attempt to establish each connection
	for i := range conns {
		if retrierOpts != nil {
			// With retry logic
			conns[i], err = Connect(uint8(retrierOpts.Count), retrierOpts.Interval, func() (T, error) {
				return connFunc()
			})
			if err != nil {
				// Return immediately on first error
				return nil, err
			}
		} else {
			// Without retry logic (single attempt)
			conns[i], err = connFunc()
			if err != nil {
				// Return immediately on first error
				return nil, err
			}
		}
	}

	// Return all successfully established connections
	return conns, nil
}
