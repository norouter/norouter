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
	"runtime"
	"testing"

	"gotest.tools/v3/assert"
)

func TestExpand(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Logf("untested on Windows")
	}
	cases := map[string]string{
		"/foo":           "",
		"~":              "",
		"~/foo":          "",
		"$HOME/foo":      "",
		"$HOMEWRONG/foo": "environment variable \"HOMEWRONG\" is unset",
		"~root/foo":      "unsupported form",
	}
	for s, expectedErr := range cases {
		got, err := Expand(s)
		if expectedErr == "" {
			t.Logf("Expand(%q) = %q", s, got)
			assert.NilError(t, err)
		} else {
			assert.ErrorContains(t, err, expectedErr)
		}
	}
}
