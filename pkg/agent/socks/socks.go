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

package socks

import (
	"context"
	"fmt"
	"net"

	"github.com/cybozu-go/usocksd/socks"
	"github.com/norouter/norouter/pkg/agent/resolver"
	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/adapters/gonet"
	"gvisor.dev/gvisor/pkg/tcpip/network/ipv4"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
)

func NewServer(st *stack.Stack, rv *resolver.Resolver) (*socks.Server, error) {
	d, err := NewDialer(st, rv)
	if err != nil {
		return nil, err
	}
	s := &socks.Server{
		Dialer:      d,
		SilenceLogs: true,
	}
	return s, nil
}

func NewDialer(st *stack.Stack, rv *resolver.Resolver) (socks.Dialer, error) {
	d := &dialer{
		stack:    st,
		resolver: rv,
	}
	return d, nil
}

type dialer struct {
	stack    *stack.Stack
	resolver *resolver.Resolver
}

func (d *dialer) Dial(req *socks.Request) (net.Conn, error) {
	s := req.Hostname
	if s == "" && req.IP != nil {
		s = req.IP.String()
	}
	if !d.resolver.Interesting(s) {
		addr := fmt.Sprintf("%s:%d", s, req.Port)
		return net.Dial("tcp", addr)
	}
	gonetIP, err := d.resolver.Resolve(s)
	if err != nil {
		return nil, err
	}
	fullAddr := tcpip.FullAddress{
		Addr: tcpip.Address(gonetIP),
		Port: uint16(req.Port),
	}
	return gonet.DialContextTCP(context.TODO(), d.stack, fullAddr, ipv4.ProtocolNumber)
}
