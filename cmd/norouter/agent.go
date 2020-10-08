/*
   Copyright (C) NoRouter authors.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package main

import (
	"encoding/json"
	"os"

	"github.com/norouter/norouter/pkg/agent"
	"github.com/norouter/norouter/pkg/stream/jsonmsg"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

var agentCommand = &cli.Command{
	Name:   "agent",
	Usage:  "agent (No need to launch manually)",
	Action: agentAction,
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:   "automated",
			Usage:  "Must be always specified. This hidden flag just exists for printing \"do not launch agent manually\" error",
			Hidden: true,
		},
		&cli.StringFlag{
			Name:   "debug-init-config",
			Usage:  "Start the agent with initial ConfigureRequestArgs. Should be used only for debugging and testing.",
			Hidden: true,
		},
	},
}

func agentAction(clicontext *cli.Context) error {
	if !clicontext.Bool("automated") {
		return errors.New("do not launch agent manually")
	}
	initConfig, err := loadInitConfig(clicontext)
	if err != nil {
		return err
	}
	a, err := agent.New(os.Stdout, os.Stdin, initConfig)
	if err != nil {
		return err
	}
	return a.Run()
}

func loadInitConfig(clicontext *cli.Context) (*jsonmsg.ConfigureRequestArgs, error) {
	s := clicontext.String("debug-init-config")
	if s == "" {
		return nil, nil
	}
	var x jsonmsg.ConfigureRequestArgs
	if err := json.Unmarshal([]byte(s), &x); err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal \"debug-init-config\" string as ConfigureRequestArgs: %q", s)
	}
	return &x, nil
}
