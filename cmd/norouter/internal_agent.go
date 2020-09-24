package main

import (
	"os"

	"github.com/norouter/norouter/pkg/agent"
	"github.com/norouter/norouter/pkg/agent/config"
	"github.com/urfave/cli/v2"
)

var internalAgentCommand = &cli.Command{
	Name:   "agent",
	Usage:  "agent",
	Action: internalAgentAction,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "me",
			Usage:    "my virtual IP without port and proto, e.g. \"127.0.42.101\"",
			Required: true,
		},
		&cli.StringSliceFlag{
			Name:  "other",
			Usage: "other virtual IP, port, and optionally proto, e.g. \"127.0.42.102:8080/tcp\"",
		},
		&cli.StringSliceFlag{
			Name:  "forward",
			Usage: "local forward, e.g. \"8080:127.0.0.1:80/tcp\"",
		},
	},
}

func internalAgentAction(clicontext *cli.Context) error {
	me, err := config.ParseMe(clicontext.String("me"))
	if err != nil {
		return err
	}
	var others []*config.Other
	for _, s := range clicontext.StringSlice("other") {
		o, err := config.ParseOther(s)
		if err != nil {
			return err
		}
		others = append(others, o)
	}
	var forwards []*config.Forward
	for _, s := range clicontext.StringSlice("forward") {
		f, err := config.ParseForward(s)
		if err != nil {
			return err
		}
		forwards = append(forwards, f)
	}
	a, err := agent.New(me, others, forwards, os.Stdout, os.Stdin)
	if err != nil {
		return err
	}
	return a.Run()
}
