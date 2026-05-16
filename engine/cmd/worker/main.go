package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/anirudh/argus/engine/internal/argus"
)

func main() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	errChan := make(chan error, 1)
	go func() {
		errChan <- argus.RunWorker()
	}()

	select {
	case sig := <-sigChan:
		fmt.Printf("\nReceived signal: %v — shutting down gracefully...\n", sig)
		argus.ShutdownWorker()
		<-errChan
		fmt.Println("Worker stopped")

	case err := <-errChan:
		if err != nil {
			fmt.Fprintln(os.Stderr, "worker error:", err)
			os.Exit(1)
		}
	}
}
