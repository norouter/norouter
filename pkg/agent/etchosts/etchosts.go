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
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/sirupsen/logrus"
)

const (
	MarkerBegin = "<Added-by-NoRouter>"
	MarkerEnd   = "</Added-by-NoRouter>"
)

// SystemEtcHostsFilePath returns "/etc/hosts" on Unix, "%SystemRoot\\System32\\drivers\\etc\\hosts" on Windows
func SystemEtcHostsFilePath() (string, error) {
	if runtime.GOOS != "windows" {
		return "/etc/hosts", nil
	}
	systemRoot, ok := os.LookupEnv("SystemRoot")
	if !ok {
		return "", errors.New("failed to get SystemRoot")
	}
	// systemRoot is typically "C:\WINDOWS"
	s := filepath.Join(systemRoot, "System32", "drivers", "etc", "hosts")
	return s, nil
}

// Populate populates hosts file.
//
// When filePath is empty, it is interpreted as "/etc/hosts" on Unix, "%SystemRoot\\System32\\drivers\\etc\\hosts" on Windows
//
// A backup file is created when backupFileSuffix is specified and the backup file does not exist.
func Populate(filePath string, hostnameMap map[string]net.IP, backupFileSuffix string) error {
	sys, err := SystemEtcHostsFilePath()
	if err != nil {
		return err
	}
	if filePath == "" {
		filePath = sys
	}

	sysR, err := os.Open(sys)
	if err != nil {
		logrus.WithError(err).Warnf("failed to read %q", sys)
		sysR = nil
	}
	b := populate(hostnameMap, sysR)
	sysR.Close()

	if backupFileSuffix != "" {
		backupFilePath := filePath + backupFileSuffix
		if err := backupFile(backupFilePath, filePath); err != nil {
			logrus.WithError(err).Warnf("failed to backup %q as %q", filePath, backupFilePath)
		}
	}

	w, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer w.Close()
	if _, err := w.Write(b); err != nil {
		return err
	}
	return nil
}

func backupFile(backupFilePath, filePath string) error {
	if _, err := os.Stat(backupFilePath); err == nil {
		logrus.Debugf("backup file %q already exists, skipping creating a backup file", backupFilePath)
		return nil
	}
	logrus.Debugf("creating a backup of %q as %q", filePath, backupFilePath)
	backup, err := os.Create(backupFilePath)
	if err != nil {
		return err
	}
	defer backup.Close()
	orig, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer orig.Close()
	_, err = io.Copy(backup, orig)
	return err
}

func populate(hostnameMap map[string]net.IP, systemEtcHosts io.Reader) []byte {
	var b bytes.Buffer
	if systemEtcHosts != nil {
		if err := readButSkipMarkedRegion(&b, systemEtcHosts); err != nil {
			logrus.WithError(err).Warn("failed to read existing file")
		}
	}
	fmt.Fprintf(&b, "# %s\n", MarkerBegin)
	for name, ip := range hostnameMap {
		fmt.Fprintf(&b, "%s %s\n", ip.String(), name)
	}
	fmt.Fprintf(&b, "# %s\n", MarkerEnd)
	return b.Bytes()
}

// readButSkipMarkedRegion skips the <Added-By-NoRouter> </Added-By-NoRouter> region
func readButSkipMarkedRegion(w io.Writer, r io.Reader) error {
	scanner := bufio.NewScanner(r)
	skip := false
	for scanner.Scan() {
		line := scanner.Text()
		sawMarkerEnd := false
		if strings.HasPrefix(line, "#") {
			com := strings.TrimSpace(line[1:])
			switch com {
			case MarkerBegin:
				skip = true
			case MarkerEnd:
				sawMarkerEnd = true
			}
		}
		if !skip {
			fmt.Fprintln(w, line)
		}
		if sawMarkerEnd {
			skip = false
		}
	}
	return scanner.Err()
}
