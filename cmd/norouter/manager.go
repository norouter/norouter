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
	"io/ioutil"
	"os"
	"strings"

	"github.com/norouter/norouter/pkg/manager"
	"github.com/norouter/norouter/pkg/manager/manifest"
	"github.com/norouter/norouter/pkg/manager/manifest/parsed"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v2"
)

var managerCommand = &cli.Command{
	Name:      "manager",
	Aliases:   []string{"m"},
	Usage:     "manager (default subcommand)",
	ArgsUsage: "[FILE]",
	Action:    managerAction,
}

func managerAction(clicontext *cli.Context) error {
	manifestPath := clicontext.Args().First()
	if manifestPath == "" {
		return errors.Errorf("no manifest file path was specified, run `%s show-example` to show an example", os.Args[0])
	}
	parsed, err := loadManifest(manifestPath)
	if err != nil {
		return err
	}
	ccSet, err := manager.NewCmdClientSet(parsed)
	if err != nil {
		return err
	}
	for vip, client := range ccSet.ByVIP {
		logrus.Debugf("client for %q: %q", vip, client.String())
	}
	m, err := manager.New(ccSet)
	if err != nil {
		return err
	}
	return m.Run()
}

func loadManifest(filePath string) (*parsed.ParsedManifest, error) {
	b, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	var raw manifest.Manifest
	if err := yaml.Unmarshal(b, &raw); err != nil {
		if strings.Contains(err.Error(), "found character that cannot start any token") {
			err = errors.Wrap(err, "failed to parse YAML, maybe you are mixing up tabs and spaces? YAML does not allow tabs.")
		}
		return nil, err
	}
	return parsed.New(&raw)
}
