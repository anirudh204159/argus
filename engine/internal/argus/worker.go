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

var workerCancel context.CancelFunc

func RunWorker() error {
	if err := initRedis(); err != nil {
		return fmt.Errorf("redis init: %w", err)
	}
	defer rdb.Close()
	fmt.Println("Worker connected to Redis")

	// Start metrics HTTP server (port 9102 for worker)
	go serveMetrics(9102)

	if err := initMetadataDB(); err != nil {
		return fmt.Errorf("metadata db init: %w", err)
	}
	defer metadataDB.Close()
	fmt.Println("Worker connected to metadata DB")

	ctx, cancel := context.WithCancel(context.Background())
	workerCancel = cancel
	defer cancel()

	go refreshSubscriptions(ctx)
	go watchConfigChanges(ctx)

	time.Sleep(500 * time.Millisecond)

	err := rdb.XGroupCreateMkStream(ctx, streamKey, workerConsumerGroup, "$").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		return fmt.Errorf("create consumer group: %w", err)
	}

	fmt.Printf("Worker '%s' listening on stream '%s' (group: %s)\n",
		workerConsumerName, streamKey, workerConsumerGroup)

	for {
		if ctx.Err() != nil {
			fmt.Println("Worker shutdown complete")
			return nil
		}

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
			if ctx.Err() != nil {
				fmt.Println("Worker shutdown complete")
				return nil
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

func ShutdownWorker() {
	if workerCancel != nil {
		workerCancel()
	}
}

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
			dlqWritesTotal.Inc()
			fmt.Printf("    ! sub %d → DLQ after %d attempts: %s\n",
				sub.ID, attempts, result.ErrorMessage)
		}
	}

	return nil
}

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

func contains(list []string, target string) bool {
	for _, item := range list {
		if item == target {
			return true
		}
	}
	return false
}
