package main

import (
	"os"

	"github.com/norfrt6-lab/go-dev-cli/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
