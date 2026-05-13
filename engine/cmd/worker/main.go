package main

import (
	"fmt"
	"os"

	"github.com/anirudh/argus/engine/internal/argus"
)

func main() {
	if err := argus.RunWorker(); err != nil {
		fmt.Fprintln(os.Stderr, "worker error:", err)
		os.Exit(1)
	}
}
