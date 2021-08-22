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

package resolver

import (
	"context"
	"encoding/json"
	"fmt"
	"net"

	"github.com/miekg/dns"
	"github.com/norouter/norouter/pkg/router"
	"github.com/norouter/norouter/pkg/stream"
	"github.com/norouter/norouter/pkg/stream/jsonmsg"

	"github.com/sirupsen/logrus"
	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/adapters/gonet"
	"gvisor.dev/gvisor/pkg/tcpip/network/ipv4"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
)

func New(hostnameMap map[string]net.IP, routes []jsonmsg.Route, vips []net.IP, st *stack.Stack, nameServers []jsonmsg.NameServer, eventSender *stream.Sender) (*Resolver, error) {
	rt, err := router.New(routes, vips)
	if err != nil {
		return nil, err
	}
	canonMap := make(map[string]net.IP)
	for k, v := range hostnameMap {
		canonMap[dns.CanonicalName(k)] = v
	}
	r := &Resolver{
		router:      rt,
		canonMap:    canonMap,
		stack:       st,
		nameServers: nameServers,
		eventSender: eventSender,
	}
	return r, nil
}

type Resolver struct {
	router      *router.Router
	canonMap    map[string]net.IP
	stack       *stack.Stack
	nameServers []jsonmsg.NameServer
	eventSender *stream.Sender
}

// Interesting returns true if req shouldn't be passed through to the OS.
// i.e. the req should be dialed with gonet dial.
// req must be either hostname or IP
func (r *Resolver) Interesting(req string) bool {
	reqAsIP := net.ParseIP(req)
	reqCanon := dns.CanonicalName(req)
	// The actual router is in manager.
	// In agent, we only check whether it is in the routes config or not
	routeRes := r.router.Route(reqAsIP)
	routeWithHostnameRes := r.router.RouteWithHostname(reqCanon)
	for canon, ip := range r.canonMap {
		if reqCanon == canon {
			return true
		}
		if ip.Equal(reqAsIP) {
			return true
		}
		if ip.Equal(routeRes) {
			return true
		}
		if ip.Equal(routeWithHostnameRes) {
			return true
		}
	}

	// if req is a hostname, try resolve it, and see whether the resolved IP
	// is interesting.
	if reqAsIP == nil {
		if lookedUp, err := net.LookupIP(req); err == nil {
			for _, ip := range lookedUp {
				if x := r.Interesting(ip.String()); x {
					return true
				}
			}
		}
	}
	return false
}

// Resolve must be called only when r.Interesting() returned true.
// Behavior of Resolve is undefined when r.Interesteing() returned false.
func (r *Resolver) Resolve(req string) (net.IP, error) {
	if reqAsIP := net.ParseIP(req); reqAsIP != nil {
		return reqAsIP, nil
	}
	reqCanon := dns.CanonicalName(req)
	for canon, ip := range r.canonMap {
		if reqCanon == canon {
			return ip, nil
		}
	}
	routeWithHostnameRes := r.router.RouteWithHostname(reqCanon)
	if routeWithHostnameRes == nil {
		lookedUp, err := net.LookupIP(req)
		if err != nil {
			return nil, err
		}
		for _, f := range lookedUp {
			if f = f.To4(); f != nil {
				return f, nil
			}
		}
		return nil, fmt.Errorf("failed to resolve %q", req)
	}
	for _, ns := range r.nameServers {
		if ns.IP.Equal(routeWithHostnameRes) && ns.Proto == "tcp" {
			res, err := resolveWithGonetTCP(r.stack, req, ns.IP, ns.Port)
			if err != nil {
				return nil, err
			}
			r.router.Learn(res, routeWithHostnameRes, true)
			routeSuggestion := jsonmsg.RouteSuggestionEventData{
				IP:    res,
				Route: routeWithHostnameRes,
			}
			if err := sendRouteSuggestionEvent(r.eventSender, &routeSuggestion); err != nil {
				logrus.WithError(err).Warn("failed to send RouteSuggestion event")
			}
			// TODO: shuffle?
			return res[0], nil
		}
	}
	return nil, fmt.Errorf("no gonet DNS found for %q", req)
}

func resolveWithGonetTCP(st *stack.Stack, query string, srv net.IP, port uint16) ([]net.IP, error) {
	fullAddr := tcpip.FullAddress{
		Addr: tcpip.Address(srv),
		Port: port,
	}
	conn, err := gonet.DialContextTCP(context.TODO(), st, fullAddr, ipv4.ProtocolNumber)
	if err != nil {
		return nil, err
	}
	dnsConn := &dns.Conn{
		Conn: conn,
	}
	client := &dns.Client{
		Net: "tcp",
	}
	req := &dns.Msg{
		MsgHdr: dns.MsgHdr{
			Opcode:           dns.OpcodeQuery,
			Id:               dns.Id(),
			RecursionDesired: true,
		},
		Question: []dns.Question{
			{
				Name:   dns.Fqdn(query),
				Qtype:  dns.TypeA,
				Qclass: dns.ClassINET,
			},
		},
	}
	reply, _, err := client.ExchangeWithConn(req, dnsConn)
	if err != nil {
		return nil, err
	}
	var res []net.IP
	for _, rr := range reply.Answer {
		if a, ok := rr.(*dns.A); ok {
			res = append(res, a.A)
		}
	}
	if len(res) == 0 {
		return nil, fmt.Errorf("failed to lookup %q with gonet DNS %s:%d/tcp: reply=%+v", query, srv.String(), port, reply)
	}
	return res, nil
}

func sendRouteSuggestionEvent(sender *stream.Sender, dat *jsonmsg.RouteSuggestionEventData) error {
	datJSON, err := json.Marshal(dat)
	if err != nil {
		return err
	}
	ev := jsonmsg.Event{
		Type: jsonmsg.EventTypeRouteSuggestion,
		Data: datJSON,
	}
	evJSON, err := json.Marshal(ev)
	if err != nil {
		return err
	}
	msg := jsonmsg.Message{
		Type: jsonmsg.TypeEvent,
		Body: evJSON,
	}
	msgJSON, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	pkt := &stream.Packet{
		Type:    stream.TypeJSON,
		Payload: msgJSON,
	}
	return sender.Send(pkt)
}
