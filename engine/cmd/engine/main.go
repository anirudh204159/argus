package main

import (
	"fmt"
	"os"

	"github.com/anirudh/argus/engine/internal/argus"
)

func main() {
	if err := argus.RunEngine(); err != nil {
		fmt.Fprintln(os.Stderr, "engine error:", err)
		os.Exit(1)
	}
}
