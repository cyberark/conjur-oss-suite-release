package main

import (
	"github.com/cyberark/conjur-oss-suite-release/pkg/cli"
	"github.com/cyberark/conjur-oss-suite-release/pkg/log"
)

func main() {
	log.OutLogger.Printf("Starting changelog parser...")

	options := cli.Options{}

	err := options.HandleInput()
	if err != nil {
		log.ErrLogger.Fatal(err)
	}

	err = cli.RunParser(options)
	if err != nil {
		log.ErrLogger.Fatal(err)
	}
}
