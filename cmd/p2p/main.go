package main

import (
	"os"

	"github.com/dimchansky/go-p2p-forwarding/cmd/p2p/commands"
	"github.com/jessevdk/go-flags"
)

func main() {
	p := flags.NewParser(&commands.Root, flags.Default)

	_, err := p.Parse()
	handleError(err)
}

func handleError(err error) {
	if err != nil {
		if flagsErr, ok := err.(*flags.Error); ok {
			if flagsErr.Type == flags.ErrHelp {
				os.Exit(0)
			}
		}

		os.Exit(1)
	}
}
