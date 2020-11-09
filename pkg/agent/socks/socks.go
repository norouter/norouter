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
	"github.com/norouter/norouter/pkg/router"
	"github.com/pkg/errors"
	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/adapters/gonet"
	"gvisor.dev/gvisor/pkg/tcpip/network/ipv4"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
)

func NewServer(st *stack.Stack, hostnameMap map[string]net.IP, router *router.Router) (*socks.Server, error) {
	d, err := NewDialer(st, hostnameMap, router)
	if err != nil {
		return nil, err
	}
	s := &socks.Server{
		Dialer:      d,
		SilenceLogs: true,
	}
	return s, nil
}

func NewDialer(st *stack.Stack, hostnameMap map[string]net.IP, router *router.Router) (socks.Dialer, error) {
	d := &dialer{
		stack:       st,
		hostnameMap: hostnameMap,
		router:      router,
	}
	return d, nil
}

type dialer struct {
	stack       *stack.Stack
	hostnameMap map[string]net.IP
	router      *router.Router
}

func (d *dialer) Dial(req *socks.Request) (net.Conn, error) {
	gonetDial := false
	ip := req.IP
	if len(ip) == 0 {
		if req.Hostname != "" {
			if parsed := net.ParseIP(req.Hostname); len(parsed) != 0 {
				ip = parsed
			}
			if resolved, ok := d.hostnameMap[req.Hostname]; ok {
				ip = resolved
				gonetDial = true
			} else {
				if lookedUp, err := net.LookupIP(req.Hostname); err == nil {
					for _, f := range lookedUp {
						if f = f.To4(); f != nil {
							ip = f
						}
					}
				}
			}
		}
	}

	if len(ip) == 0 {
		reqWithoutPassword := *req
		if reqWithoutPassword.Password != "" {
			reqWithoutPassword.Password = "********"
		}
		return nil, errors.Errorf("failed to determine IP for request %+v", reqWithoutPassword)
	}

	if !d.router.Route(ip).Equal(ip) {
		gonetDial = true
	}

	if gonetDial {
		fullAddr := tcpip.FullAddress{
			Addr: tcpip.Address(ip),
			Port: uint16(req.Port),
		}
		return gonet.DialContextTCP(context.TODO(), d.stack, fullAddr, ipv4.ProtocolNumber)
	}

	return net.Dial("tcp", fmt.Sprintf("%s:%d", ip.String(), req.Port))
}
