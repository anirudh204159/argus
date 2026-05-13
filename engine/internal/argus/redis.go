package argus

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
)

const (
	redisAddr    = "localhost:6379"
	streamKey    = "argus:events"
	streamMaxLen = 10000 // cap stream size (oldest events dropped if exceeded)
)

var rdb *redis.Client

func initRedis() error {
	rdb = redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	// Sanity check — confirm we can talk to Redis at all
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis ping failed: %w", err)
	}
	return nil
}

// publishEvent serializes an Event to JSON and pushes it onto the Redis Stream.
func publishEvent(ctx context.Context, ev Event) error {
	payload, err := json.Marshal(ev)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	args := &redis.XAddArgs{
		Stream: streamKey,
		MaxLen: streamMaxLen,
		Approx: true, // ~MAXLEN — more efficient
		Values: map[string]any{
			"payload": payload,
		},
	}

	_, err = rdb.XAdd(ctx, args).Result()
	return err
}
