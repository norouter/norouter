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

package hostaliases

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

type FQDNBackend = string

const (
	NipIO              = FQDNBackend("nip.io")
	DefaultFQDNBackend = NipIO
)

func Populate(w io.Writer, backend FQDNBackend, hostnameMap map[string]net.IP) error {
	if backend != NipIO {
		return fmt.Errorf("unknown backend %q", backend)
	}
	m := make(map[string]string)
	if env, ok := os.LookupEnv("HOSTALIASES"); ok && env != "" {
		if err := readExistingFile(m, env); err != nil {
			logrus.WithError(err).Warnf("failed to read HOSTALIASES file %q", env)
		}
	}
	for name, ip := range hostnameMap {
		if strings.Contains(name, ".") {
			logrus.Debugf("hostaliases: ignoring %q (contains dot)", name)
			continue
		}
		ip4 := ip.To4()
		if ip4 == nil {
			logrus.Debugf("hostaliases: ignoring %q (not IPv4)", name)
			continue
		}
		ip4FQDN := fmt.Sprintf("%d.%d.%d.%d.%s", ip4[0], ip4[1], ip4[2], ip4[3], backend)
		m[name] = ip4FQDN
	}
	for k, v := range m {
		fmt.Fprintf(w, "%s %s\n", k, v)
	}
	return nil
}

func readExistingFile(m map[string]string, filePath string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) == 2 {
			k, v := fields[0], fields[1]
			m[k] = v
		}
	}
	return scanner.Err()
}
