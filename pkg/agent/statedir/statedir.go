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

package statedir

import (
	"errors"
	"io/ioutil"
	"net"
	"os"
	"os/user"
	"path/filepath"

	"github.com/norouter/norouter/pkg/agent/etchosts"
	"github.com/norouter/norouter/pkg/agent/filepathutil"
	"github.com/norouter/norouter/pkg/agent/hostaliases"
	"github.com/sirupsen/logrus"
)

const readmeMD = `# NoRouter agent state directory

This directory was created by NoRouter agent.
You can safely remove this directory if NoRouter is not running.

Files in this directory:
- README.md:    This file
- hosts:        Can be used as /etc/hosts
- hostaliases:  Can be used as $HOSTALIASES

## About "hosts" file:
The hosts file can be used as /etc/hosts if you copy to there manually.

To write /etc/hosts automatically, set .[]hosts.writeEtcHosts in the manifest file.

## About "hostaliases" file
The hostaliases file can be used as $HOSTALIASES if supported by applications.

Note that the file does not contain virtual hostnames with dots, such as "host1.norouter.local".
See hostname(7).

## Changing the path of this directory
Set .[]hosts.stateDir.pathOnAgent in the manifest file.
To disallow creating this directory, set .[]hosts.stateDir.disable to true.

- - -
For further information, see https://norouter.io/
`

// Populate populates the state dir.
// When the dir path is empty, it is interpreted as "~/.norouter/agent".
// The dir path is expanded using ../filepathutil.Expand .
//
// The following files are created in the directory: "hosts", "hostaliases", "README.md"
func Populate(dirPath string, hostnameMap map[string]net.IP) error {
	var err error
	dirPath, err = expandDirPath(dirPath)
	if err != nil {
		return err
	}
	logrus.Debugf("populating state dir %q", dirPath)
	if err = os.MkdirAll(dirPath, 0755); err != nil {
		return err
	}

	// Populate ~/.norouter/agent/hosts
	hostsFilePath := filepath.Join(dirPath, "hosts")
	if err = etchosts.Populate(hostsFilePath, hostnameMap, ""); err != nil {
		return err
	}

	// Populate ~/.norouter/agent/hostaliases
	hostAliasesFilePath := filepath.Join(dirPath, "hostaliases")
	hostAliasesFile, err := os.Create(hostAliasesFilePath)
	if err != nil {
		return err
	}
	defer hostAliasesFile.Close()
	if err = hostaliases.Populate(hostAliasesFile, hostaliases.DefaultFQDNBackend, hostnameMap); err != nil {
		return err
	}

	// Populate ~/.norouter/agent/README.md
	readmeMDFilePath := filepath.Join(dirPath, "README.md")
	if err = ioutil.WriteFile(readmeMDFilePath, []byte(readmeMD), 0644); err != nil {
		return err
	}
	return nil
}

func expandDirPath(dirPath string) (string, error) {
	if dirPath == "" {
		u, err := user.Current()
		if err != nil {
			return "", err
		}
		if u.HomeDir == "" {
			return "", errors.New("u.HomeDir is empty")
		}
		return filepath.Join(u.HomeDir, ".norouter", "agent"), nil
	}
	return filepathutil.Expand(dirPath)
}
