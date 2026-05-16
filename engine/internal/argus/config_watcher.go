package argus

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

const configChannel = "argus:config"

func watchConfigChanges(ctx context.Context) {
	pubsub := rdb.Subscribe(ctx, configChannel)
	defer pubsub.Close()

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
			subscriptionsActive.Set(float64(len(subs)))

			fmt.Printf("[config] reloaded — %d active subscriptions\n", len(subs))
		}
	}
}

// Suppress "imported and not used" linter for the redis package
// (used in the type assertion inside Channel())
var _ = redis.Nil
