package argus

import (
	"context"
	"fmt"
	"math"
	"time"
)

// deliverWithRetries attempts delivery with exponential backoff.
// Returns the final result and the number of attempts made.
func deliverWithRetries(ctx context.Context, sub Subscription, ev Event, eventID string) (deliveryResult, int) {
	maxAttempts := sub.RetryMax + 1 // initial attempt + retries
	if maxAttempts < 1 {
		maxAttempts = 1
	}

	var result deliveryResult

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		result = deliver(ctx, sub, ev, attempt)
		logDelivery(ctx, sub.ID, eventID, attempt, result)

		if result.Success {
			fmt.Printf("    ✓ delivered to sub %d (%s) in %dms (attempt %d)\n",
				sub.ID, sub.Name, result.LatencyMS, attempt)
			return result, attempt
		}

		fmt.Printf("    ✗ attempt %d/%d failed for sub %d: %s\n",
			attempt, maxAttempts, sub.ID, result.ErrorMessage)

		if attempt < maxAttempts {
			backoff := backoffDuration(attempt)
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return result, attempt
			}
		}
	}

	return result, maxAttempts
}

// backoffDuration returns the wait time before retry N (1-indexed).
// Pattern: 1s, 2s, 4s, 8s, 16s, capped at 30s.
func backoffDuration(attempt int) time.Duration {
	seconds := math.Pow(2, float64(attempt-1))
	if seconds > 30 {
		seconds = 30
	}
	return time.Duration(seconds) * time.Second
}
