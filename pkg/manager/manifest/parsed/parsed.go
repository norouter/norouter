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

package parsed

import (
	"net"
	"strconv"
	"strings"

	"github.com/norouter/norouter/pkg/manager/manifest"
	"github.com/norouter/norouter/pkg/stream/jsonmsg"
	"github.com/pkg/errors"
)

type ParsedManifest struct {
	Raw             *manifest.Manifest
	Hosts           map[string]*Host
	PublicHostPorts []*jsonmsg.IPPortProto
}

type Host struct {
	Cmd   []string
	VIP   net.IP
	Ports []*jsonmsg.Forward
}

func New(raw *manifest.Manifest) (*ParsedManifest, error) {
	pm := &ParsedManifest{
		Raw:   raw,
		Hosts: make(map[string]*Host),
	}

	uniqueVIPs := make(map[string]struct{})

	for name, rh := range raw.Hosts {
		vip := net.ParseIP(rh.VIP)
		if vip == nil {
			return nil, errors.Errorf("failed to parse virtual IP %q", rh.VIP)
		}
		vip = vip.To4()
		if vip == nil {
			return nil, errors.Errorf("failed to parse virtual IP (v4) %q", rh.VIP)
		}
		h := &Host{
			Cmd: rh.Cmd,
			VIP: vip,
		}
		for _, p := range rh.Ports {
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
		pm.Hosts[name] = h
		uniqueVIPs[rh.VIP] = struct{}{}
	}
	if len(uniqueVIPs) != len(raw.Hosts) {
		return nil, errors.Errorf("expected to have %d unique virtual IPs (VIPs), got %d", len(raw.Hosts), len(uniqueVIPs))
	}
	return pm, nil
}

// ParseForward parses --forward=8080:127.0.0.1:80[/tcp] flag
func ParseForward(forward string) (*jsonmsg.Forward, error) {
	s := strings.TrimSuffix(forward, "/tcp")
	if strings.Contains(s, "/") {
		// TODO: support "/udp" suffix
		return nil, errors.Errorf("cannot parse \"forward\" address %q", forward)
	}
	split := strings.Split(s, ":")
	if len(split) != 3 {
		return nil, errors.Errorf("cannot parse \"forward\" address %q", forward)
	}
	listenPort, err := strconv.Atoi(split[0])
	if err != nil {
		return nil, errors.Wrapf(err, "cannot parse \"forward\" address %q", forward)
	}
	connectIP := net.ParseIP(split[1])
	if connectIP == nil {
		return nil, errors.Errorf("cannot parse \"forward\" address %q", forward)
	}
	connectPort, err := strconv.Atoi(split[2])
	if err != nil {
		return nil, errors.Wrapf(err, "cannot parse \"forward\" address %q", forward)
	}
	f := &jsonmsg.Forward{
		ListenPort:  uint16(listenPort),
		ConnectIP:   connectIP,
		ConnectPort: uint16(connectPort),
		Proto:       "tcp",
	}
	return f, nil
}
