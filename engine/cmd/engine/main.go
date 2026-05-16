package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/anirudh/argus/engine/internal/argus"
)

func main() {
	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Run the engine in a goroutine so we can listen for signals in main
	errChan := make(chan error, 1)
	go func() {
		errChan <- argus.RunEngine()
	}()

	// Wait for either an error or a shutdown signal
	select {
	case sig := <-sigChan:
		fmt.Printf("\nReceived signal: %v — shutting down gracefully...\n", sig)
		argus.ShutdownEngine()
		// Give cleanup time to finish
		<-errChan
		fmt.Println("Engine stopped")

	case err := <-errChan:
		if err != nil {
			fmt.Fprintln(os.Stderr, "engine error:", err)
			os.Exit(1)
		}
	}
}
