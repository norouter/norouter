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
	"sync"

	"github.com/golang/groupcache/lru"
	"github.com/miekg/dns"
	"github.com/norouter/norouter/pkg/stream/jsonmsg"
	"github.com/pkg/errors"
	"github.com/ryanuber/go-glob"
)

func New(routes []jsonmsg.Route, reserved []net.IP) (*Router, error) {
	learntNeverForget := make(map[string]string)
	for _, ip := range reserved {
		ip = ip.To4()
		if ip == nil {
			return nil, errors.Errorf("unexpected ip %s", ip.String())
		}
		s := ip.String()
		learntNeverForget[s] = s
	}
	learntMayForget := lru.New(512)
	r := &Router{
		learntNeverForget: learntNeverForget,
		learntMayForget:   learntMayForget,
	}
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
	mu                sync.RWMutex
	learntNeverForget map[string]string
	learntMayForget   *lru.Cache
	ipEntries         []ipEntry
	globEntries       []globEntry
}

type ipEntry struct {
	IPNet net.IPNet
	Via   net.IP
}

type globEntry struct {
	Glob string
	Via  net.IP
}

func (r *Router) Learn(to []net.IP, suggestedRoute net.IP, mayForget bool) {
	suggestedRoute = suggestedRoute.To4()
	if suggestedRoute == nil {
		return
	}
	mapV := suggestedRoute.String()
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, f := range to {
		ip := f.To4()
		if ip == nil {
			continue
		}
		mapK := ip.String()
		if mayForget {
			r.learntMayForget.Add(mapK, mapV)
		} else {
			r.learntNeverForget[mapK] = mapV
		}
	}
}

// Route won't return nil (unless to is nil)
func (r *Router) Route(to net.IP) net.IP {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if ip4 := to.To4(); ip4 != nil {
		k := ip4.String()
		if v, ok := r.learntNeverForget[k]; ok {
			return net.ParseIP(v)
		}
		if lruV, ok := r.learntMayForget.Get(k); ok {
			return net.ParseIP(lruV.(string))
		}
	}

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
	r.mu.RLock()
	defer r.mu.RUnlock()

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
