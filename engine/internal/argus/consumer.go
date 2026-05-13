package argus

import (
	"context"
	"fmt"
	"strings"
)

// consumer receives events from the channel, publishes to Redis, and logs to console.
// In future phases the console print goes away; Redis becomes the only sink.
func consumer(ch <-chan Event) {
	ctx := context.Background()
	for ev := range ch {
		if err := publishEvent(ctx, ev); err != nil {
			fmt.Printf("publish error: %v (event: %s.%s/%s)\n", err, ev.Schema, ev.Table, ev.Operation)
			continue
		}
		printEvent(ev) // keep printing during dev so we can see what's flowing
	}
}

// printEvent renders an Event for the console.
func printEvent(ev Event) {
	fmt.Printf("[%s] %s.%s → published\n", ev.Operation, ev.Schema, ev.Table)
	if ev.Before != nil {
		fmt.Printf("    before: %s\n", formatMap(ev.Before))
	}
	if ev.After != nil {
		fmt.Printf("    after:  %s\n", formatMap(ev.After))
	}
}

// formatMap turns a map into a "{k: v, k: v}" string for display.
func formatMap(m map[string]interface{}) string {
	parts := []string{}
	for k, v := range m {
		parts = append(parts, fmt.Sprintf("%s: %v", k, v))
	}
	return "{" + strings.Join(parts, ", ") + "}"
}
