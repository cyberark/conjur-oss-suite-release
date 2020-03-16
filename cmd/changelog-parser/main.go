package main

import (
	"github.com/cyberark/conjur-oss-suite-release/pkg/cli"
	"github.com/cyberark/conjur-oss-suite-release/pkg/log"
)

func main() {
	log.OutLogger.Printf("Starting changelog parser...")

	options := cli.Options{}

	options.HandleInput()

	err := cli.RunParser(options)
	if err != nil {
		log.ErrLogger.Fatal(err)
	}
}
