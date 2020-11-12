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
	"bytes"
	"strings"
	"testing"

	"gotest.tools/v3/assert"
)

func TestReadButSkipMarkedRegion(t *testing.T) {

	type testCase struct {
		s string
	}

	testCases := []testCase{
		{
			s: `# no NoRouter entries
127.0.0.1       localhost
# some comment
192.168.0.100   somehost
# some comment2
192.168.0.101   somehost2
# EOF
`,
		},

		{
			s: `# with NoRouter entries
127.0.0.1       localhost
# some comment
192.168.0.100   somehost
# some comment2
# <Added-by-NoRouter>
# norouter comment
127.0.42.100 host0
127.0.42.101 host1
127.0.42.102 host2
# </Added-by-NoRouter>
192.168.0.101   somehost2
# EOF
`,
		},
	}

	for i, tc := range testCases {
		r := strings.NewReader(tc.s)
		var b bytes.Buffer
		err := readButSkipMarkedRegion(&b, r)
		assert.NilError(t, err)
		t.Logf("=== BEGIN: %d===", i)
		t.Log(b.String())
		t.Logf("=== END : %d===", i)
	}
}
