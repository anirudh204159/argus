package argus

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	workerConsumerGroup = "argus-workers"
	workerConsumerName  = "worker-1"
	workerBlockTimeout  = 5 * time.Second
)

// RunWorker consumes events from the Redis Stream and processes them.
// In Phase 8 the "process" step becomes HTTP webhook delivery with retries.
// For now, it just prints and acknowledges.
func RunWorker() error {
	if err := initRedis(); err != nil {
		return fmt.Errorf("redis init: %w", err)
	}
	defer rdb.Close()
	fmt.Println("Worker connected to Redis")

	ctx := context.Background()

	// Create the consumer group if it doesn't exist.
	// MKSTREAM tells Redis to auto-create the stream if it doesn't exist yet.
	err := rdb.XGroupCreateMkStream(ctx, streamKey, workerConsumerGroup, "$").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		return fmt.Errorf("create consumer group: %w", err)
	}

	fmt.Printf("Worker '%s' listening on stream '%s' (group: %s)\n",
		workerConsumerName, streamKey, workerConsumerGroup)

	for {
		// XREADGROUP with > means "give me messages I haven't seen yet"
		streams, err := rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    workerConsumerGroup,
			Consumer: workerConsumerName,
			Streams:  []string{streamKey, ">"},
			Count:    10,
			Block:    workerBlockTimeout,
		}).Result()

		if err != nil {
			if err == redis.Nil {
				// No new messages within the block timeout — just loop
				continue
			}
			fmt.Println("read error:", err)
			time.Sleep(1 * time.Second)
			continue
		}

		for _, stream := range streams {
			for _, msg := range stream.Messages {
				if err := handleMessage(ctx, msg); err != nil {
					fmt.Printf("handle error for %s: %v\n", msg.ID, err)
					// don't ack on error — Redis will redeliver
					continue
				}
				// Ack on success
				if err := rdb.XAck(ctx, streamKey, workerConsumerGroup, msg.ID).Err(); err != nil {
					fmt.Printf("ack error for %s: %v\n", msg.ID, err)
				}
			}
		}
	}
}

// handleMessage processes one event. For now, just decodes and prints it.
// In Phase 8 this becomes "POST to webhook with HMAC signing and retries."
func handleMessage(ctx context.Context, msg redis.XMessage) error {
	payloadRaw, ok := msg.Values["payload"].(string)
	if !ok {
		return fmt.Errorf("missing or non-string payload field")
	}

	var ev Event
	if err := json.Unmarshal([]byte(payloadRaw), &ev); err != nil {
		return fmt.Errorf("unmarshal payload: %w", err)
	}

	fmt.Printf("[worker] received %s on %s.%s (msg id: %s)\n",
		ev.Operation, ev.Schema, ev.Table, msg.ID)

	if ev.Before != nil {
		fmt.Printf("    before: %v\n", ev.Before)
	}
	if ev.After != nil {
		fmt.Printf("    after:  %v\n", ev.After)
	}

	return nil
}
