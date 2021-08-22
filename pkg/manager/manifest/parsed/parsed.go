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
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/google/shlex"
	"github.com/norouter/norouter/pkg/builtinports"
	"github.com/norouter/norouter/pkg/manager/manifest"
	"github.com/norouter/norouter/pkg/stream/jsonmsg"
)

type ParsedManifest struct {
	Raw             *manifest.Manifest
	Hosts           map[string]*Host
	PublicHostPorts []*jsonmsg.IPPortProto
	Routes          []jsonmsg.Route
	NameServers     []jsonmsg.NameServer
}

type Host struct {
	Cmd           []string
	VIP           net.IP
	Ports         []*jsonmsg.Forward
	HTTP          HTTP
	SOCKS         SOCKS
	Loopback      Loopback
	StateDir      StateDir
	Aliases       []string
	WriteEtcHosts bool
}

type HTTP struct {
	Listen string
}

type SOCKS struct {
	Listen string
}

type Loopback struct {
	Disable bool
}

type StateDir struct {
	PathOnAgent string
	Disable     bool
}

func New(raw *manifest.Manifest) (*ParsedManifest, error) {
	if ht := raw.HostTemplate; ht != nil {
		if ht.VIP != "" {
			return nil, errors.New("the HostTemplate must not have VIP")
		}
		if ht.Cmd != nil {
			return nil, errors.New("the HostTemplate must not have Cmd")
		}
		if ht.Aliases != nil {
			return nil, errors.New("the HostTemplate must not have Aliases")
		}
	}

	pm := &ParsedManifest{
		Raw:   raw,
		Hosts: make(map[string]*Host),
	}

	uniqueNames := make(map[string]struct{})
	uniqueVIPs := make(map[string]struct{})

	for name, rh := range raw.Hosts {
		if _, ok := uniqueNames[name]; ok {
			return nil, fmt.Errorf("name conflict: %q", name)
		}
		uniqueNames[name] = struct{}{}
		vip := net.ParseIP(rh.VIP)
		if vip == nil {
			return nil, fmt.Errorf("failed to parse virtual IP %q", rh.VIP)
		}
		vip = vip.To4()
		if vip == nil {
			return nil, fmt.Errorf("failed to parse virtual IP (v4) %q", rh.VIP)
		}
		h := &Host{
			VIP: vip,
		}
		cmd, err := ParseCmd(rh.Cmd)
		if err != nil {
			return nil, err
		}
		h.Cmd = cmd

		rawPorts := rh.Ports
		if raw.HostTemplate != nil {
			rawPorts = append(raw.HostTemplate.Ports, rh.Ports...)
		}
		for _, p := range rawPorts {
			f, err := ParseForward(p)
			if err != nil {
				return nil, err
			}
			h.Ports = append(h.Ports, f)
			pm.PublicHostPorts = append(pm.PublicHostPorts,
				&jsonmsg.IPPortProto{
					IP:    vip,
					Port:  f.ListenPort,
					Proto: f.Proto,
				})
		}
		if raw.HostTemplate != nil {
			if raw.HostTemplate.HTTP != nil {
				h.HTTP.Listen = raw.HostTemplate.HTTP.Listen
			}
			if raw.HostTemplate.SOCKS != nil {
				h.SOCKS.Listen = raw.HostTemplate.SOCKS.Listen
			}
			if raw.HostTemplate.Loopback != nil {
				h.Loopback.Disable = raw.HostTemplate.Loopback.Disable
			}
			if raw.HostTemplate.StateDir != nil {
				h.StateDir.PathOnAgent = raw.HostTemplate.StateDir.PathOnAgent
				h.StateDir.Disable = raw.HostTemplate.StateDir.Disable
			}
			if raw.HostTemplate.WriteEtcHosts != nil {
				h.WriteEtcHosts = *raw.HostTemplate.WriteEtcHosts
			}
		}
		if rh.HTTP != nil {
			h.HTTP.Listen = rh.HTTP.Listen
		}
		if rh.SOCKS != nil {
			h.SOCKS.Listen = rh.SOCKS.Listen
		}
		if rh.Loopback != nil {
			h.Loopback.Disable = rh.Loopback.Disable
		}
		if rh.StateDir != nil {
			h.StateDir.PathOnAgent = rh.StateDir.PathOnAgent
			h.StateDir.Disable = rh.StateDir.Disable
		}
		if rh.WriteEtcHosts != nil {
			h.WriteEtcHosts = *rh.WriteEtcHosts
		}
		for _, a := range rh.Aliases {
			if _, ok := uniqueNames[a]; ok {
				return nil, fmt.Errorf("name conflict: %q", a)
			}
			uniqueNames[a] = struct{}{}
			h.Aliases = append(h.Aliases, a)
		}

		pm.Hosts[name] = h
		uniqueVIPs[rh.VIP] = struct{}{}
	}
	if len(uniqueVIPs) != len(raw.Hosts) {
		return nil, fmt.Errorf("expected to have %d unique virtual IPs (VIPs), got %d", len(raw.Hosts), len(uniqueVIPs))
	}
	for _, rawRoute := range raw.Routes {
		route, err := parseRoute(rawRoute, pm.Hosts)
		if err != nil {
			return nil, err
		}
		pm.Routes = append(pm.Routes, *route)
	}

	// TODO: support specifying custom DNS ports via YAML
	for _, h := range pm.Hosts {
		ns := jsonmsg.NameServer{
			IPPortProto: jsonmsg.IPPortProto{
				IP:    h.VIP,
				Port:  builtinports.DNSTCP,
				Proto: "tcp",
			},
		}
		pm.NameServers = append(pm.NameServers, ns)
	}
	return pm, nil
}

func parseRoute(raw manifest.Route, hosts map[string]*Host) (*jsonmsg.Route, error) {
	r := &jsonmsg.Route{}
	if h, ok := hosts[raw.Via]; ok {
		r.Via = h.VIP
	} else {
		ip := net.ParseIP(raw.Via)
		if ip == nil {
			return nil, fmt.Errorf("failed to parse \"via\" IP: %q", raw.Via)
		}
		ip = ip.To4()
		if ip == nil {
			return nil, fmt.Errorf("failed to parse \"via\" IPv4: %q", raw.Via)
		}
		r.Via = ip
	}
	for _, rawTo := range raw.To {
		_, _, err := net.ParseCIDR(rawTo)
		if err == nil {
			r.ToCIDR = append(r.ToCIDR, rawTo)
		} else {
			if net.ParseIP(rawTo) != nil {
				return nil, fmt.Errorf("expected CIDR or hostname glob, got unexpected IP %q, maybe you forgot to add \"/32\" suffix?", rawTo)
			}
			r.ToHostnameGlob = append(r.ToHostnameGlob, rawTo)
		}
	}
	return r, nil
}

func ParseCmd(cmdX interface{}) ([]string, error) {
	switch cmd := cmdX.(type) {
	case []string:
		return cmd, nil
	case []interface{}:
		var ret []string
		for _, x := range cmd {
			s, ok := x.(string)
			if !ok {
				return nil, fmt.Errorf("expected cmd to be []string, got %+v", cmd)
			}
			ret = append(ret, s)
		}
		return ret, nil
	case string:
		split, err := shlex.Split(cmd)
		if err != nil {
			return nil, fmt.Errorf("failed to split cmd string %q: %w", cmd, err)
		}
		return split, nil
	case nil:
		return nil, nil
	default:
		return nil, fmt.Errorf("expected cmd to be either []string or string, got %+T (%+v)", cmd, cmd)
	}
}

// ParseForward parses "8080:127.0.0.1:80[/tcp]"
func ParseForward(forward string) (*jsonmsg.Forward, error) {
	s := strings.TrimSuffix(forward, "/tcp")
	if strings.Contains(s, "/") {
		// TODO: support "/udp" suffix
		return nil, fmt.Errorf("cannot parse \"forward\" address %q", forward)
	}
	split := strings.Split(s, ":")
	if len(split) != 3 {
		return nil, fmt.Errorf("cannot parse \"forward\" address %q", forward)
	}
	listenPort, err := strconv.Atoi(split[0])
	if err != nil {
		return nil, fmt.Errorf("cannot parse \"forward\" address %q: %w", forward, err)
	}
	connectIP := split[1]
	connectPort, err := strconv.Atoi(split[2])
	if err != nil {
		return nil, fmt.Errorf("cannot parse \"forward\" address %q: %w", forward, err)
	}
	f := &jsonmsg.Forward{
		ListenPort:  uint16(listenPort),
		ConnectIP:   connectIP,
		ConnectPort: uint16(connectPort),
		Proto:       "tcp",
	}
	return f, nil
}
