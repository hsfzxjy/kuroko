package cmd

import (
	"kuroko-linux/internal/log"

	"github.com/jessevdk/go-flags"
)

type Options struct {
	Verbose []bool `short:"v" long:"verbose" description:"Verbose output"`
}

var CLIOptions = Options{}
var CLIParser = flags.NewParser(&CLIOptions, flags.HelpFlag|flags.PassDoubleDash)
var logger = log.Name("cmd")
