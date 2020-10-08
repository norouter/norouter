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

package parsed

import (
	"net"
	"testing"

	"github.com/norouter/norouter/pkg/manager/manifest"
	"github.com/norouter/norouter/pkg/stream/jsonmsg"
	"gopkg.in/yaml.v2"
	"gotest.tools/v3/assert"
)

func TestNew(t *testing.T) {
	type testCase struct {
		s             string
		expectedError string
	}

	testCases := []testCase{
		{
			s: `# valid manifest
hosts:
  foo:
    vip: "127.0.42.100"
  bar:
    cmd: ["docker", "exec", "-i", "foo", "norouter"]
    vip: "127.0.42.101"
    ports: ["8080:127.0.0.1:80"]
  baz:
    cmd: ["docker", "exec", "-i", "bar", "norouter"]
    vip: 127.0.42.102
    ports:
    - 8080:127.0.0.1:80
`,
		},
		{
			s: `# invalid manifest with overlapping VIPs
hosts:
  foo:
    vip: "127.0.42.100"
  bar:
    cmd: ["docker", "exec", "-i", "foo", "norouter"]
    vip: "127.0.42.101"
  baz:
    cmd: ["docker", "exec", "-i", "bar", "norouter"]
    vip: "127.0.42.101"
    ports: ["8080:127.0.0.1:80"]
`,
			expectedError: "expected to have 3 unique virtual IPs (VIPs)",
		},
	}

	for _, c := range testCases {
		var raw manifest.Manifest
		if err := yaml.Unmarshal([]byte(c.s), &raw); err != nil {
			t.Fatal(err)
		}
		_, err := New(&raw)
		if c.expectedError != "" {
			assert.ErrorContains(t, err, c.expectedError)
			continue
		}
		assert.NilError(t, err)
	}

}

func TestParseForward(t *testing.T) {
	type testCase struct {
		s             string
		expectedError string
		expected      jsonmsg.Forward
	}

	testCases := []testCase{
		{
			s: "8080:127.0.0.1:80/tcp",
			expected: jsonmsg.Forward{
				ListenPort:  8080,
				ConnectIP:   net.ParseIP("127.0.0.1"),
				ConnectPort: 80,
				Proto:       "tcp",
			},
		},
		{
			s: "8081:192.168.1.2:81",
			expected: jsonmsg.Forward{
				ListenPort:  8081,
				ConnectIP:   net.ParseIP("192.168.1.2"),
				ConnectPort: 81,
				Proto:       "tcp",
			},
		},
		{
			s:             "8080:127.0.0.1:80/udp",
			expectedError: "cannot parse",
		},
		{
			s:             "8080",
			expectedError: "cannot parse",
		},
	}

	for _, c := range testCases {
		f, err := ParseForward(c.s)
		if c.expectedError != "" {
			assert.ErrorContains(t, err, c.expectedError)
			continue
		}
		assert.NilError(t, err)
		assert.DeepEqual(t, c.expected, *f)
	}
}
