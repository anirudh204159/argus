package main

import (
	"fmt"
	"strings"
)

// consumer receives events from the channel and prints them.
// In future phases this will be replaced with "send to Redis" or "POST to webhook".
func consumer(ch <-chan Event) {
	for ev := range ch {
		printEvent(ev)
	}
}

// printEvent renders an Event for the console.
func printEvent(ev Event) {
	fmt.Printf("[%s] %s.%s\n", ev.Operation, ev.Schema, ev.Table)
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
