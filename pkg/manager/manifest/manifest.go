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
	//
	// HostTemplate MUST NOT contain the following fields:
	// VIP, Cmd, and Aliases:
	//
	// HostTemplate can be specified since NoRouter v0.4.0
	HostTemplate *Host `yaml:"hostTemplate,omitempty"`

	// Host defines hosts.
	//
	// The key string is used as the virtual hostname that can
	// be resolved using HOSTALIASES, HTTP proxy, or SOCKS proxy.
	//
	// The virtual hostname string SHOULD NOT contain dot symbols.
	// The virtual hostnames with dot symbols are not added to HOSTALIASES file.
	Hosts map[string]Host `yaml:"hosts"`

	// Routes specifies routes to access hosts behind bastions.
	// Routes only makes sense for HTTP and SOCKS proxy modes.
	// Routes is optional.
	// Routes can be specified since NoRouter v0.4.0
	Routes []Route `yaml:"routes",omitempty`
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
	// The example above forwards connections to the TCP port 8080
	// of the virtual IP (e.g. 127.0.42.101) to the TCP port 80
	// of the real IP 127.0.0.1.
	//
	// Currently, only TCP protocol is supported.
	//
	// Ports are optional.
	//
	// Ports are appended to HostTemplate.Ports
	// when HostTemplate is specified.
	Ports []string `yaml:"ports,omitempty"`

	// HTTP can be specified since NoRouter v0.4.0
	HTTP *HTTP `yaml:"http,omitempty"`

	// SOCKS can be specified since NoRouter v0.4.0
	SOCKS *SOCKS `yaml:"socks,omitempty"`

	// Loopback can be specified since NoRouter v0.4.0
	Loopback *Loopback `yaml:"loopback,omitempty"`

	// StateDir can be specified since NoRouter v0.4.0
	StateDir *StateDir `yaml:"stateDir,omitempty"`

	// Aliases specify aliases of the virtual hostname.
	// e.g. ["nginx.example.com", "nginx"]
	// Aliases may contain dot symbols, but aliases with dot symbols are not added to HOSTALIASES file.
	//
	// Aliases can be specified since NoRouter v0.4.0
	Aliases []string `yaml:"aliases",omitempty`

	// WriteEtcHosts specifies to write /etc/hosts when possible.
	// WriteEtcHosts is expected to be used with Docker and Kubernetes containers.
	//
	// WriteEtcHosts can be specified since NoRouter v0.4.0
	WriteEtcHosts *bool `yaml:"writeEtcHosts",omitempty`
}

// HTTP can be specified since NoRouter v0.4.0
type HTTP struct {
	// Listen specifies an address of HTTP proxy to be listened by NoRouter agent processes.
	// The address is typically a local address, e.g. "127.0.0.1:18080".
	// When the address is not specified, HTTP proxy is disabled.
	Listen string `yaml:"listen,omitempty"`
}

// SOCKS can be specified since NoRouter v0.4.0
type SOCKS struct {
	// Listen specifies an address of SOCKS proxy to be listened by NoRouter agent processes.
	// The address is typically a local address, e.g. "127.0.0.1:18081".
	// When the address is not specified, SOCKS proxy is disabled.
	//
	// Supported protocol versions: SOCKS4, SOCKS4a, and SOCKS5
	Listen string `yaml:"listen,omitempty"`
}

// Loopback can be specified since NoRouter v0.4.0
type Loopback struct {
	// Disable disables listening on multi-loopback addresses such as 127.0.42.100, 127.0.42.101...
	//
	// When Disable is set, HTTP.Listen should be specified to enable HTTP proxy.
	Disable bool `yaml:"disable,omitempty"`
}

// StateDir can be specified since NoRouter v0.4.0
type StateDir struct {
	// PathOnAgent specifies the state directory path on the agent.
	//
	// When PathOnAgent is not set, the path is set to "~/.norouter/agent".
	// The path string can contain "~" and "${ENVVAR}".
	// Env vars are resolved on the agent, not on the manager.
	//
	// PathOnAgent is ignored when Disable is set.
	PathOnAgent string `yaml:"pathOnAgent,omitempty"`

	// Disable disables creating the state directory.
	Disable bool `yaml:"disable,omitempty"`
}

// Route can be specified since NoRouter v0.4.0.
// Route only makes sense for HTTP and SOCKS proxy modes.
type Route struct {
	// To must be IPv4 CIDR or hostname globs
	// e.g. 0.0.0.0/0 (all IPs), 192.168.95.0/24, 192.168.95.100/32, *.cloud1.example.com
	To []string `yaml:"to"`

	// TODO: support "NotTo"

	// Via is a bastion.
	// Via is a virtual hostname or a virtual IP.
	Via string `yaml:"via"`
}
