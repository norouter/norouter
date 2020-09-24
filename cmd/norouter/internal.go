package main

import (
	"github.com/urfave/cli/v2"
)

var internalCommand = &cli.Command{
	Name:   "internal",
	Usage:  "Internal commands",
	Hidden: true,
	Subcommands: []*cli.Command{
		internalAgentCommand,
	},
}
