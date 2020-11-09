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

// Package router might be a misnomer :(
package router

import (
	"net"

	"github.com/norouter/norouter/pkg/stream/jsonmsg"
)

func New(routes []jsonmsg.Route) (*Router, error) {
	r := &Router{}
	for _, msg := range routes {
		for _, to := range msg.To {
			_, ipnet, err := net.ParseCIDR(to)
			if err != nil {
				return nil, err
			}
			e := entry{IPNet: *ipnet, Via: msg.Via}
			r.entries = append(r.entries, e)
		}
	}
	return r, nil
}

type Router struct {
	entries []entry
}

type entry struct {
	IPNet net.IPNet
	Via   net.IP
}

func (r *Router) Route(to net.IP) net.IP {
	// reverse order
	for i := len(r.entries) - 1; i >= 0; i-- {
		e := r.entries[i]
		if e.IPNet.Contains(to) {
			return e.Via
		}
	}
	return to
}
