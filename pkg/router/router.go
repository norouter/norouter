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

	"github.com/miekg/dns"
	"github.com/norouter/norouter/pkg/stream/jsonmsg"
	"github.com/ryanuber/go-glob"
)

func New(routes []jsonmsg.Route) (*Router, error) {
	r := &Router{}
	for _, msg := range routes {
		for _, to := range msg.ToCIDR {
			_, ipnet, err := net.ParseCIDR(to)
			if err != nil {
				return nil, err
			}
			e := ipEntry{IPNet: *ipnet, Via: msg.Via}
			r.ipEntries = append(r.ipEntries, e)
		}
		for _, to := range msg.ToHostnameGlob {
			e := globEntry{Glob: to, Via: msg.Via}
			r.globEntries = append(r.globEntries, e)
		}

	}
	return r, nil
}

type Router struct {
	ipEntries   []ipEntry
	globEntries []globEntry
}

type ipEntry struct {
	IPNet net.IPNet
	Via   net.IP
}

type globEntry struct {
	Glob string
	Via  net.IP
}

// Route won't return nil (unless to is nil)
func (r *Router) Route(to net.IP) net.IP {
	// reverse order
	for i := len(r.ipEntries) - 1; i >= 0; i-- {
		e := r.ipEntries[i]
		if e.IPNet.Contains(to) {
			return e.Via
		}
	}
	return to
}

// RouteWithHostname may return nil
func (r *Router) RouteWithHostname(hostname string) net.IP {
	canon := dns.CanonicalName(hostname)
	// reverse order
	for i := len(r.globEntries) - 1; i >= 0; i-- {
		e := r.globEntries[i]
		if glob.Glob(dns.CanonicalName(e.Glob), canon) {
			return e.Via
		}
	}
	return nil
}
