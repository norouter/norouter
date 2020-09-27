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

package manifest

type Manifest struct {
	Hosts map[string]Host `yaml:"hosts"`
}

type Host struct {
	Cmd   []string `yaml:"cmd"`   // e.g. ["docker", "exec", "-i", "host1", "norouter"]
	VIP   string   `yaml:"vip"`   // e.g. "127.0.42.101"
	Ports []string `yaml:"ports"` // e.g. ["8080:127.0.0.1:80"], or ["8080:127.0.0.1:80/tcp"]
}
