package main

import (
	"os"

	"github.com/futuretea/kubectl-yadt/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
