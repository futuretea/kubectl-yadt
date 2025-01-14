package main

import (
	"log"

	"github.com/ibuildthecloud/wtfk8s/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
