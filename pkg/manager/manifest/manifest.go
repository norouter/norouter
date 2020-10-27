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

package manifest

type Manifest struct {
	// HostTemplate is optional.
	// HostTemplate must not contain VIP and Cmd.
	HostTemplate *Host           `yaml:"hostTemplate"`
	Hosts        map[string]Host `yaml:"hosts"`
}

type Host struct {
	// VIP is a virtual IP address.
	// Currently, only IPv4 addresses are supported.
	//
	// e.g. "127.0.42.101"
	//
	// VIP must be always specified.
	VIP string `yaml:"vip"` // e.g. "127.0.42.101"

	// Cmd is either []string or string.
	//
	// e.g. ["docker", "exec", "-i", "host1", "norouter"]
	// e.g. "docker exec -i host1 norouter"
	//
	// Cmd is optional.
	Cmd interface{} `yaml:"cmd,omitempty"`

	// Ports specify port forwarding.
	//
	// e.g. ["8080:127.0.0.1:80"]
	// e.g. ["8080:127.0.0.1:80/tcp"]
	//
	// Ports are optional.
	//
	// Ports are appended to HostTemplate.Ports
	// when HostTemplate is specified.
	Ports []string `yaml:"ports,omitempty"`

	HTTP *HTTP `yaml:"http,omitempty"`
}

type HTTP struct {
	Listen string `yaml:"listen,omitempty"`
}
