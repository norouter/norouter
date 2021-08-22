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

package filepathutil

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
)

// Expand expands "~/foo" and "$HOME/user" to "/home/user/foo"
func Expand(s string) (string, error) {
	var err error
	envExpander := func(x string) string {
		y, ok := os.LookupEnv(x)
		if !ok {
			err = fmt.Errorf("failed to expand %q: environment variable %q is unset", s, x)
		}
		return y
	}
	s = os.Expand(s, envExpander)
	if err != nil {
		return "", err
	}
	if strings.HasPrefix(s, "~") {
		u, err := user.Current()
		if err != nil {
			return "", fmt.Errorf("failed to expand %q: %w", s, err)
		}
		if u.HomeDir == "" {
			return "", fmt.Errorf("failed to expand %q: home dir is empty", s)
		}
		if s == "~" {
			return u.HomeDir, nil
		}
		if strings.HasPrefix(s, "~/") || (runtime.GOOS == "windows" && strings.HasPrefix(s, "~\\")) {
			res := filepath.Join(u.HomeDir, s[1:])
			return res, nil
		}
		// otherwise like "~username/foo"
		return "", fmt.Errorf("unsupported form: %q", s)
	}
	return s, nil
}
