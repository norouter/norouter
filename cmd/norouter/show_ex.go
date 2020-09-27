/*
   Copyright (C) Nippon Telegraph and Telephone Corporation.

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
	"fmt"
	"strings"

	"github.com/urfave/cli/v2"
)

var showExampleCommand = &cli.Command{
	Name:    "show-example",
	Aliases: []string{"show-ex"},
	Usage:   "show an example manifest",
	Action:  showExampleAction,
}

func exampleManifest() string {
	s := `# Example manifest for NoRouter.
# Run @BACKQUOTE@norouter <FILE>@BACKQUOTE@ to start NoRouter with the specified manifest file.
#
# The @BACKQUOTE@norouter@BACKQUOTE@ binary needs to be installed on all the remote hosts.
# Run @BACKQUOTE@norouter show-installer@BACKQUOTE@ to show the installation script.
#
hosts:
# localhost
  local:
    vip: "127.0.42.100"
# Docker container (docker exec)
  docker:
    cmd: ["docker", "exec", "-i", "some-container", "norouter"]
    vip: "127.0.42.101"
    ports: ["8080:127.0.0.1:80"]
# Podman container (podman exec)
  podman:
    cmd: ["podman", "exec", "-i", "some-container", "norouter"]
    vip: "127.0.42.102"
    ports: ["8080:127.0.0.1:80"]
# Kubernetes Pod (kubectl exec)
  kube:
    cmd: ["kubectl", "--context=some-context", "exec", "-i", "some-pod", "--", "norouter"]
    vip: "127.0.42.103"
    ports: ["8080:127.0.0.1:80"]
# SSH
# If your key has a passphrase, make sure to configure ssh-agent so that NoRouter can login to the remote host automatically.
  ssh:
    cmd: ["ssh", "some-user@some-ssh-host.example.com", "--", "norouter"]
    vip: "127.0.42.104"
    ports: ["8080:127.0.0.1:80"]
`
	s = strings.ReplaceAll(s, "@BACKQUOTE@", "`")
	return s
}

func showExampleAction(clicontext *cli.Context) error {
	fmt.Print(exampleManifest())
	return nil
}
