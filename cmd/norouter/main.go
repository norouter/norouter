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
	"os"
	"strings"

	"github.com/norouter/norouter/pkg/version"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func main() {
	logrus.SetFormatter(newLogrusFormatter())
	if err := newApp().Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}

func newApp() *cli.App {
	debug := false
	app := cli.NewApp()
	app.Name = "norouter"
	app.Usage = "the easiest multi-host & multi-cloud networking ever. No root privilege is required."
	app.Version = version.Version
	app.Description = strings.ReplaceAll(`
  NoRouter is the easiest multi-host & multi-cloud networking ever.
  And yet, NoRouter does not require any privilege such as @BACKQUOTE@sudo@BACKQUOTE@ or @BACKQUOTE@docker run --privileged@BACKQUOTE@.

  NoRouter implements unprivileged networking by using multiple loopback addresses such as 127.0.42.101 and 127.0.42.102.
  The hosts in the network are connected by forwarding packets over stdio streams like @BACKQUOTE@ssh@BACKQUOTE@, @BACKQUOTE@docker exec@BACKQUOTE@, @BACKQUOTE@podman exec@BACKQUOTE@, @BACKQUOTE@kubectl exec@BACKQUOTE@, and whatever.

  Quick usage:
  - Install the @BACKQUOTE@norouter@BACKQUOTE@ binary to all the hosts. Run @BACKQUOTE@norouter show-installer@BACKQUOTE@ to show an installation script.
  - Create a manifest YAML file. Run @BACKQUOTE@norouter show-example@BACKQUOTE@ to show an example manifest.
  - Run @BACKQUOTE@norouter <FILE>@BACKQUOTE@ to start NoRouter with the specified manifest YAML file.

  Web site: https://norouter.io`, "@BACKQUOTE@", "`")

	app.Flags = []cli.Flag{
		&cli.BoolFlag{
			Name:        "debug",
			Usage:       "debug mode",
			Destination: &debug,
		},
	}
	app.Flags = append(app.Flags, managerFlags...)
	app.Before = func(context *cli.Context) error {
		if debug {
			logrus.SetLevel(logrus.DebugLevel)
		}
		return nil
	}
	app.Commands = []*cli.Command{
		managerCommand,
		agentCommand,
		showExampleCommand,
		showInstallerCommand,
	}
	app.Action = managerAction
	return app
}

func newLogrusFormatter() logrus.Formatter {
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "<unknown>"
	}
	return &logrusFormatter{
		prefix:    hostname + ": ",
		Formatter: &logrus.TextFormatter{},
	}
}

type logrusFormatter struct {
	prefix string
	logrus.Formatter
}

func (lf *logrusFormatter) Format(e *logrus.Entry) ([]byte, error) {
	b, err := lf.Formatter.Format(e)
	if err != nil {
		return b, err
	}
	return append([]byte(lf.prefix), b...), nil
}
