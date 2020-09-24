package main

import (
	"os"

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
	app.Usage = "NoRouter: the easiest multi-host & multi-cloud networking ever. No root privilege required."

	app.Flags = []cli.Flag{
		&cli.BoolFlag{
			Name:        "debug",
			Usage:       "debug mode",
			Destination: &debug,
		},
	}
	app.Before = func(context *cli.Context) error {
		if debug {
			logrus.SetLevel(logrus.DebugLevel)
		}
		return nil
	}
	app.Commands = []*cli.Command{
		routerCommand,
		internalCommand,
	}
	app.Action = routerAction
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
