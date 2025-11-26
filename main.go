package main

import (
	"os"

	"github.com/typicalfo/netgaze/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
