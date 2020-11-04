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

package loopback

import (
	"context"
	"fmt"
	"net"
	"runtime"
	"strings"
	"syscall"

	"github.com/norouter/norouter/pkg/agent/bicopy/bicopyutil"
	"github.com/norouter/norouter/pkg/stream/jsonmsg"
	"github.com/pkg/errors"
	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/adapters/gonet"
	"gvisor.dev/gvisor/pkg/tcpip/network/ipv4"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
)

func isBSD(goos string) bool {
	if strings.Contains(goos, "bsd") {
		return true
	}
	switch goos {
	case "darwin", "dragonfly":
		return true
	}
	return false
}

func listen(proto, addr string) (net.Listener, error) {
	l, err := net.Listen(proto, addr)
	if err != nil {
		// "listen tcp 127.0.43.101:8080: bind: can't assign requested address"
		if errors.Is(err, syscall.EADDRNOTAVAIL) || strings.Contains(err.Error(), "can't assign requested address") {
			if isBSD(runtime.GOOS) {
				err = errors.Wrap(err, "hint: try running `sudo ifconfig lo0 alias <IP>`")
			}
		}
	}
	return l, err
}

// GoOther forwards connections to "others" VIP such as 127.0.42.102:8080, 127.0.42.103:8080..
// to the netstack network.
func GoOther(st *stack.Stack, o jsonmsg.IPPortProto) error {
	if o.Proto != "tcp" {
		return errors.Errorf("expected proto be \"tcp\", got %q", o.Proto)
	}
	oAddr := fmt.Sprintf("%s:%d", o.IP.String(), o.Port)
	l, err := listen(o.Proto, oAddr)
	if err != nil {
		return errors.Wrapf(err, "failed to listen on %q", oAddr)
	}
	dial := func(proto, addr string) (net.Conn, error) {
		if proto != "tcp" || addr != oAddr {
			return nil, errors.Errorf("expected (\"tcp\", %q), got (%q, %q))", oAddr, proto, addr)
		}
		fullAddr := tcpip.FullAddress{
			Addr: tcpip.Address(o.IP),
			Port: o.Port,
		}
		return gonet.DialContextTCP(context.TODO(), st, fullAddr, ipv4.ProtocolNumber)
	}
	go bicopyutil.BicopyAcceptDial(l, o.Proto, oAddr, dial)
	return nil
}

// GoLocalForward forwards connections to "my" VIP such as 127.0.42.101:8080
// to the underlying application such as 127.0.0.1:80
func GoLocalForward(me net.IP, f jsonmsg.Forward) error {
	if f.Proto != "tcp" {
		return errors.Errorf("expected proto be \"tcp\", got %q", f.Proto)
	}
	lh := fmt.Sprintf("%s:%d", me.String(), f.ListenPort)
	l, err := listen(f.Proto, lh)
	if err != nil {
		return errors.Wrapf(err, "failed to listen on %q", lh)
	}
	go bicopyutil.BicopyAcceptDial(l, f.Proto, fmt.Sprintf("%s:%d", f.ConnectIP, f.ConnectPort), net.Dial)
	return nil
}
