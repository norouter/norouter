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
	"net"
	"os"
	"os/user"
	"path/filepath"

	"github.com/norouter/norouter/pkg/agent/etchosts"
	"github.com/norouter/norouter/pkg/agent/filepathutil"
	"github.com/norouter/norouter/pkg/agent/hostaliases"
	"github.com/sirupsen/logrus"
)

// Populate populates the state dir.
// When the dir path is empty, it is interpreted as "~/.norouter/agent".
// The dir path is expanded using ../filepathutil.Expand .
//
// The following files are created in the directory: "hosts", "hostaliases"
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
	hostsFile, err := os.Create(hostsFilePath)
	if err != nil {
		return err
	}
	defer hostsFile.Close()
	if err = etchosts.Populate(hostsFile, hostnameMap); err != nil {
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
