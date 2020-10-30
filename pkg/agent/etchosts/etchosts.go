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

package etchosts

import (
	"fmt"
	"io"
	"net"
	"os"

	"github.com/sirupsen/logrus"
)

const marker = "Added-by-NoRouter"

func Populate(w io.Writer, hostnameMap map[string]net.IP) error {
	if err := readExistingFile(w, "/etc/hosts"); err != nil {
		logrus.WithError(err).Warn("failed to read /etc/hosts")
	}
	fmt.Fprintf(w, "# <%s>\n", marker)
	for name, ip := range hostnameMap {
		fmt.Fprintf(w, "%s %s\n", ip.String(), name)
	}
	fmt.Fprintf(w, "# </%s>\n", marker)
	return nil
}

func readExistingFile(w io.Writer, filePath string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(w, f)
	return err
}
