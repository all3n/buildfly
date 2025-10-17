package main

import (
	"log"
	"os"

	"buildfly/cmd/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		log.Fatalf("Error: %v", err)
		os.Exit(1)
	}
}
