package argus

import (
	"context"
	"fmt"
)

const configChannel = "argus:config"

// watchConfigChanges subscribes to Redis pub/sub and triggers immediate
// subscription reloads when a message arrives. Runs as a goroutine for
// the lifetime of the worker.
func watchConfigChanges(ctx context.Context) {
	pubsub := rdb.Subscribe(ctx, configChannel)
	defer pubsub.Close()

	// Block until subscription is confirmed
	if _, err := pubsub.Receive(ctx); err != nil {
		fmt.Printf("config watcher subscribe error: %v\n", err)
		return
	}

	fmt.Printf("Config watcher listening on channel '%s'\n", configChannel)

	ch := pubsub.Channel()
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}
			fmt.Printf("[config] received: %s — reloading subscriptions\n", msg.Payload)

			subs, err := loadSubscriptions(ctx)
			if err != nil {
				fmt.Printf("[config] reload error: %v\n", err)
				continue
			}

			subscriptionsMu.Lock()
			cachedSubscriptions = subs
			subscriptionsMu.Unlock()

			fmt.Printf("[config] reloaded — %d active subscriptions\n", len(subs))
		}
	}
}
