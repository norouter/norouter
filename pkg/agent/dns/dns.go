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

package dns

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"net"
	"os/exec"
	"runtime"
	"strings"

	"github.com/miekg/dns"

	"github.com/sirupsen/logrus"
	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/adapters/gonet"
	"gvisor.dev/gvisor/pkg/tcpip/network/ipv4"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
)

func New(st *stack.Stack, vip net.IP, tcpPort int, hostnameMap map[string]net.IP) (*dns.Server, error) {
	h, err := NewHandler(hostnameMap)
	if err != nil {
		return nil, err
	}
	fullAddr := tcpip.FullAddress{
		Addr: tcpip.Address(vip),
		Port: uint16(tcpPort),
	}
	l, err := gonet.ListenTCP(st, fullAddr, ipv4.ProtocolNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on %q: %w", fullAddr, err)
	}
	srv := &dns.Server{
		Handler:  h,
		Listener: l,
	}
	return srv, nil
}

func NewClientConfig() (*dns.ClientConfig, error) {
	if runtime.GOOS == "windows" {
		return newClientConfigWindows()
	}
	return dns.ClientConfigFromFile("/etc/resolv.conf")
}

func newClientConfigWindows() (*dns.ClientConfig, error) {
	powershell, err := exec.LookPath("powershell.exe")
	if err != nil {
		return nil, err
	}
	args := []string{"-NoProfile", "-NonInteractive", "Get-DnsClientServerAddress -AddressFamily IPv4 | Select-Object -ExpandProperty ServerAddresses"}
	logrus.Debugf("executing %v", append([]string{powershell}, args...))
	out, err := exec.Command(powershell, args...).Output()
	if err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(bytes.NewReader(out))
	var ips []net.IP
	for scanner.Scan() {
		line := scanner.Text()
		ip := net.ParseIP(line)
		if ip == nil {
			logrus.Warnf("unexpected line from Powershell output: %q", line)
			continue
		}
		ips = append(ips, ip)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to parse Powershell output: %w", err)
	}
	if len(ips) == 0 {
		return nil, errors.New("no DNS found")
	}
	return NewStaticClientConfig(ips)
}

func NewStaticClientConfig(ips []net.IP) (*dns.ClientConfig, error) {
	s := ``
	for _, ip := range ips {
		s += fmt.Sprintf("nameserver %s\n", ip.String())
	}
	r := strings.NewReader(s)
	return dns.ClientConfigFromReader(r)
}

func NewHandler(hostnameMap map[string]net.IP) (dns.Handler, error) {
	cc, err := NewClientConfig()
	if err != nil {
		fallbackIPs := []net.IP{net.ParseIP("8.8.8.8"), net.ParseIP("1.1.1.1")}
		logrus.WithError(err).Warnf("failed to detect system DNS, falling back to %v", fallbackIPs)
		cc, err = NewStaticClientConfig(fallbackIPs)
		if err != nil {
			return nil, err
		}
	}
	clients := []*dns.Client{
		&dns.Client{}, // UDP
		&dns.Client{Net: "tcp"},
	}
	canonMap := make(map[string]net.IP)
	for vague, ip := range hostnameMap {
		canon := dns.CanonicalName(vague)
		canonMap[canon] = ip
	}
	h := &Handler{
		clientConfig: cc,
		clients:      clients,
		canonMap:     canonMap,
	}
	return h, nil
}

type Handler struct {
	clientConfig *dns.ClientConfig
	clients      []*dns.Client
	canonMap     map[string]net.IP
}

func (h *Handler) handleQuery(w dns.ResponseWriter, req *dns.Msg) {
	var (
		reply   dns.Msg
		handled bool
	)
	reply.SetReply(req)
	for _, q := range reply.Question {
		canon := dns.CanonicalName(q.Name)
		switch q.Qtype {
		case dns.TypeA:
			if ip, ok := h.canonMap[canon]; ok {
				a := &dns.A{
					Hdr: dns.RR_Header{
						Name:   q.Name,
						Rrtype: dns.TypeA,
						Class:  dns.ClassINET,
					},
					A: ip,
				}
				reply.Answer = append(reply.Answer, a)
				handled = true
			}
		}
	}
	if handled {
		w.WriteMsg(&reply)
		return
	}
	h.handleDefault(w, req)
}

func (h *Handler) handleDefault(w dns.ResponseWriter, req *dns.Msg) {
	for _, client := range h.clients {
		for _, srv := range h.clientConfig.Servers {
			addr := fmt.Sprintf("%s:%s", srv, h.clientConfig.Port)
			reply, _, err := client.Exchange(req, addr)
			if err == nil {
				w.WriteMsg(reply)
				return
			}
		}
	}
	var reply dns.Msg
	reply.SetReply(req)
	w.WriteMsg(&reply)
}

func (h *Handler) ServeDNS(w dns.ResponseWriter, req *dns.Msg) {
	switch req.Opcode {
	case dns.OpcodeQuery:
		h.handleQuery(w, req)
	default:
		h.handleDefault(w, req)
	}
}
