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
	"strconv"
	"strings"

	"github.com/google/shlex"
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
	Cmd      []string
	VIP      net.IP
	Ports    []*jsonmsg.Forward
	HTTP     HTTP
	SOCKS    SOCKS
	Loopback Loopback
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

func New(raw *manifest.Manifest) (*ParsedManifest, error) {
	if ht := raw.HostTemplate; ht != nil {
		if ht.VIP != "" {
			return nil, errors.New("the HostTemplate must not have VIP")
		}
		if ht.Cmd != nil {
			return nil, errors.New("the HostTemplate must not have Cmd")
		}
	}

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

		pm.Hosts[name] = h
		uniqueVIPs[rh.VIP] = struct{}{}
	}
	if len(uniqueVIPs) != len(raw.Hosts) {
		return nil, errors.Errorf("expected to have %d unique virtual IPs (VIPs), got %d", len(raw.Hosts), len(uniqueVIPs))
	}
	return pm, nil
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
				return nil, errors.Errorf("expected cmd to be []string, got %+v", cmd)
			}
			ret = append(ret, s)
		}
		return ret, nil
	case string:
		split, err := shlex.Split(cmd)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to split cmd string %q", cmd)
		}
		return split, nil
	case nil:
		return nil, nil
	default:
		return nil, errors.Errorf("expected cmd to be either []string or string, got %+T (%+v)", cmd, cmd)
	}
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
