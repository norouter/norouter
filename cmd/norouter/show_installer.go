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
	"text/template"

	"github.com/norouter/norouter/pkg/version"
	"github.com/urfave/cli/v2"
)

var showInstallerCommand = &cli.Command{
	Name:   "show-installer",
	Usage:  "show script for installing NoRouter to other hosts",
	Action: showInstallerAction,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "version",
			Value: version.LatestOfficialStableRelease,
		},
	},
}

func showInstallerAction(clicontext *cli.Context) error {
	tmpl := `#!/bin/sh
set -eux
# Installation script for NoRouter
# NOTE: make sure to use the same version across all the hosts.
version="{{.Version}}"
bindir="$HOME/bin"
mkdir -p "${bindir}"
rm -f "${bindir}/norouter"
curl -o "${bindir}/norouter" --fail -L https://github.com/norouter/norouter/releases/download/v${version}/norouter-$(uname -s)-$(uname -m)
chmod +x "${bindir}/norouter"
echo "Successfully installed ${bindir}/norouter (version ${version})"
`
	m := map[string]string{
		"Version": clicontext.String("version"),
	}
	x, err := template.New("").Parse(tmpl)
	if err != nil {
		return err
	}
	return x.Execute(os.Stdout, m)
}
