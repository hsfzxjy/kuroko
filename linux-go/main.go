package main

import (
	"os"

	"kuroko-linux/cmd"
	"kuroko-linux/internal/log"

	"github.com/jessevdk/go-flags"
)

func main() {
	if _, err := cmd.CLIParser.Parse(); err != nil {
		switch flagsErr := err.(type) {
		case flags.ErrorType:
			if flagsErr == flags.ErrHelp {
				os.Exit(0)
			}
			os.Exit(1)
		default:
			log.Name("cmd").Error(flagsErr)
			os.Exit(1)
		}
	}
}
