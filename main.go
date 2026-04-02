package main

import (
	"os"

	"github.com/divinedev111/hookshot/cmd"
)

func main() {
	if err := cmd.NewRoot().Execute(); err != nil {
		os.Exit(1)
	}
}
