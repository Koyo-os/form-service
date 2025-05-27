package retrier

import "time"

// Connect attempts to establish a connection with retry logic.
//
// This generic function executes a connection function multiple times if it fails,
// waiting a specified duration between attempts. It's useful for handling temporary
// connection failures in distributed systems or unreliable networks.
//
// Type Parameters:
//   - T: The type of the connection object to be returned
//
// Parameters:
//   - retry: Maximum number of retry attempts (0 means exactly one attempt)
//   - sleep: Delay between retries in seconds (ignored if retry is 0)
//   - connector: Function that establishes the connection (returns T and error)
//
// Returns:
//   - T: The successfully established connection (on success)
//   - error: The last error encountered if all attempts failed, or nil on success
//
// Behavior:
//   - Attempts the connection up to retry+1 times (initial attempt + retries)
//   - Returns immediately on first successful connection
//   - Sleeps between failed attempts (except after the last attempt)
//   - Returns the last error if all attempts fail
//   - Zero retry value results in exactly one attempt with no waiting
//
// Example Usage:
//
//	dbConn, err := retrier.Connect(3, 2, func() (*sql.DB, error) {
//	    return sql.Open("postgres", connStr)
//	})
func Connect[T any](retry uint8, sleep uint, connector func() (T, error)) (T, error) {
	var (
		out T     // Will hold the successful connection
		err error // Will hold any connection error
	)

	// Attempt connection up to 'retry' times (total attempts = retry + 1)
	for range retry {
		out, err = connector()

		// Return immediately if connection succeeds
		if err == nil {
			return out, nil
		}

		// Wait before next attempt, except after the final attempt
		time.Sleep(time.Duration(sleep) * time.Second)
	}

	// Return either:
	// - The successful connection and nil error (unlikely in this path)
	// - The last failed connection attempt and its error
	return out, err
}
