package main

import (
	"io/ioutil"

	"github.com/norouter/norouter/pkg/router"
	"github.com/norouter/norouter/pkg/router/config"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v2"
)

var routerCommand = &cli.Command{
	Name:    "router",
	Aliases: []string{"r"},
	Usage:   "router (default subcommand)",
	Action:  routerAction,
}

func routerAction(clicontext *cli.Context) error {
	configPath := clicontext.Args().First()
	if configPath == "" {
		return errors.New("no config file path was specified")
	}
	cfg, err := loadConfig(configPath)
	if err != nil {
		return err
	}
	logrus.Debugf("config: %+v", cfg)
	ccSet, err := router.NewCmdClientSet(cfg)
	if err != nil {
		return err
	}
	for vip, client := range ccSet.ByVIP {
		logrus.Debugf("client for %q: %q", vip, client.String())
	}
	r, err := router.New(ccSet)
	if err != nil {
		return err
	}
	return r.Run()
}

func loadConfig(configPath string) (*config.Config, error) {
	configB, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	var cfg config.Config
	if err := yaml.Unmarshal(configB, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
