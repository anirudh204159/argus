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

// RunWorker consumes events from Redis and delivers them to subscribed webhooks.
func RunWorker() error {
	if err := initRedis(); err != nil {
		return fmt.Errorf("redis init: %w", err)
	}
	defer rdb.Close()
	fmt.Println("Worker connected to Redis")

	if err := initMetadataDB(); err != nil {
		return fmt.Errorf("metadata db init: %w", err)
	}
	defer metadataDB.Close()
	fmt.Println("Worker connected to metadata DB")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start background subscription refresher
	go refreshSubscriptions(ctx)

	// Give it a beat to do the first load before we start processing
	time.Sleep(500 * time.Millisecond)

	// Create the consumer group if it doesn't exist
	err := rdb.XGroupCreateMkStream(ctx, streamKey, workerConsumerGroup, "$").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		return fmt.Errorf("create consumer group: %w", err)
	}

	fmt.Printf("Worker '%s' listening on stream '%s' (group: %s)\n",
		workerConsumerName, streamKey, workerConsumerGroup)

	for {
		streams, err := rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    workerConsumerGroup,
			Consumer: workerConsumerName,
			Streams:  []string{streamKey, ">"},
			Count:    10,
			Block:    workerBlockTimeout,
		}).Result()

		if err != nil {
			if err == redis.Nil {
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
					continue
				}
				if err := rdb.XAck(ctx, streamKey, workerConsumerGroup, msg.ID).Err(); err != nil {
					fmt.Printf("ack error for %s: %v\n", msg.ID, err)
				}
			}
		}
	}
}

// handleMessage decodes an event and finds matching subscriptions.
func handleMessage(ctx context.Context, msg redis.XMessage) error {
	payloadRaw, ok := msg.Values["payload"].(string)
	if !ok {
		return fmt.Errorf("missing or non-string payload field")
	}

	var ev Event
	if err := json.Unmarshal([]byte(payloadRaw), &ev); err != nil {
		return fmt.Errorf("unmarshal payload: %w", err)
	}

	matches := matchSubscriptions(ev)
	if len(matches) == 0 {
		return nil
	}

	fmt.Printf("[worker] event %s.%s/%s → %d subscription(s)\n",
		ev.Schema, ev.Table, ev.Operation, len(matches))

	for _, sub := range matches {
		result, attempts := deliverWithRetries(ctx, sub, ev, msg.ID)
		if !result.Success {
			writeToDLQ(ctx, sub.ID, msg.ID, ev, attempts, result.ErrorMessage)
			fmt.Printf("    ! sub %d → DLQ after %d attempts: %s\n",
				sub.ID, attempts, result.ErrorMessage)
		}
	}

	return nil
}

// matchSubscriptions returns subscriptions that care about this event
// based on table name and operation type.
func matchSubscriptions(ev Event) []Subscription {
	var matches []Subscription
	for _, sub := range getSubscriptions() {
		if !contains(sub.Tables, ev.Table) {
			continue
		}
		if !contains(sub.Operations, ev.Operation) {
			continue
		}
		matches = append(matches, sub)
	}
	return matches
}

// contains is a small helper: true if `target` is in `list`.
func contains(list []string, target string) bool {
	for _, item := range list {
		if item == target {
			return true
		}
	}
	return false
}
