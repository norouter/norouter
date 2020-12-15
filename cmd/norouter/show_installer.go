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
set -eu
# Installation script for NoRouter
# NOTE: make sure to use the same version across all the hosts.
version="{{.Version}}"
echo "# Version: ${version}"

bindir="$HOME/bin"
echo "# Destination: ${bindir}/norouter"
mkdir -p "${bindir}"

download(){
	local="$1"
	remote="$2"
	echo "# Downloading ${remote}"
	if command -v curl >/dev/null 2>&1; then
		( set -x; curl -fL -o "$local" "$remote" )
	elif command -v wget >/dev/null 2>&1; then
		( set -x; wget -O "$local" "$remote" )
	else
		echo >&2 "curl or wget needs to be installed"
		exit 1
	fi
}

fname="norouter-$(uname -s)-$(uname -m).tgz"
tmp=$(mktemp -d)
download "${tmp}/${fname}" "https://github.com/norouter/norouter/releases/download/v${version}/${fname}"
download "${tmp}/SHA256SUMS" "https://github.com/norouter/norouter/releases/download/v${version}/SHA256SUMS"

if command -v sha256sum &> /dev/null; then
	(
		cd "${tmp}"
		echo "# Printing sha256sum of SHA256SUMS file itself"
		sha256sum SHA256SUMS
		echo "# Checking SHA256SUMS"
		grep "${fname}" SHA256SUMS | sha256sum -c -
		echo "# Extracting norouter executable"
		tar xzvf "${fname}"
	)
fi
if [ -x "${bindir}/norouter" ]; then
	echo "# Removing existing ${bindir}/norouter"
	rm -f "${bindir}/norouter"
fi
echo "# Installing ${tmp}/norouter onto ${bindir}/norouter"
mv "${tmp}/norouter" "${bindir}/norouter"

rm -rf "${tmp}"

echo "# Successfully installed ${bindir}/norouter (version ${version})"
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
