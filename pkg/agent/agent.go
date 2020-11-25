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

package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"

	"github.com/miekg/dns"
	"github.com/norouter/norouter/pkg/agent/bicopy"
	"github.com/norouter/norouter/pkg/agent/bicopy/bicopyutil"
	agentdns "github.com/norouter/norouter/pkg/agent/dns"
	"github.com/norouter/norouter/pkg/agent/etchosts"
	agenthttp "github.com/norouter/norouter/pkg/agent/http"
	"github.com/norouter/norouter/pkg/agent/loopback"
	"github.com/norouter/norouter/pkg/agent/netstackutil"
	"github.com/norouter/norouter/pkg/agent/netstackutil/gonetutil"
	"github.com/norouter/norouter/pkg/agent/resolver"
	agentsocks "github.com/norouter/norouter/pkg/agent/socks"
	"github.com/norouter/norouter/pkg/agent/statedir"
	"github.com/norouter/norouter/pkg/stream"
	"github.com/norouter/norouter/pkg/stream/jsonmsg"
	"github.com/norouter/norouter/pkg/version"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/adapters/gonet"
	"gvisor.dev/gvisor/pkg/tcpip/buffer"
	"gvisor.dev/gvisor/pkg/tcpip/header"
	"gvisor.dev/gvisor/pkg/tcpip/header/parse"
	"gvisor.dev/gvisor/pkg/tcpip/link/channel"
	"gvisor.dev/gvisor/pkg/tcpip/network/ipv4"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
	"gvisor.dev/gvisor/pkg/tcpip/transport/tcp"
)

const (
	mtu         = 65536
	rcvBufferSz = 2 * 1024 * 1024
	sndBufferSz = 2 * 1024 * 1024
	chanSz      = 256
)

func newStack() *stack.Stack {
	opts := stack.Options{
		NetworkProtocols:   []stack.NetworkProtocolFactory{ipv4.NewProtocol},
		TransportProtocols: []stack.TransportProtocolFactory{tcp.NewProtocol},
		HandleLocal:        false,
	}
	st := stack.New(opts)
	st.SetForwarding(ipv4.ProtocolNumber, false)
	st.SetTransportProtocolOption(tcp.ProtocolNumber,
		&tcpip.TCPReceiveBufferSizeRangeOption{
			Min:     4096,
			Default: rcvBufferSz,
			Max:     rcvBufferSz,
		})
	st.SetTransportProtocolOption(tcp.ProtocolNumber,
		&tcpip.TCPSendBufferSizeRangeOption{
			Min:     4096,
			Default: sndBufferSz,
			Max:     sndBufferSz,
		})
	tcpSACK := tcpip.TCPSACKEnabled(true)
	st.SetTransportProtocolOption(tcp.ProtocolNumber,
		&tcpSACK)
	tcpDelay := tcpip.TCPDelayEnabled(true)
	st.SetTransportProtocolOption(tcp.ProtocolNumber,
		&tcpDelay)
	return st
}

func New(w io.Writer, r io.Reader, initConfig *jsonmsg.ConfigureRequestArgs) (*Agent, error) {
	a := &Agent{
		stack: newStack(),
		sender: &stream.Sender{
			Writer: w,
		},
		receiver: &stream.Receiver{
			Reader: r,
		},
		routeHooks: make(map[uint64]*routeHook),
	}
	if initConfig != nil {
		logrus.Debugf("using init config %+v", initConfig)
		if err := a.configure(initConfig); err != nil {
			return nil, err
		}
	}
	return a, nil
}

type Agent struct {
	stack        *stack.Stack
	sender       *stream.Sender
	receiver     *stream.Receiver
	config       *jsonmsg.ConfigureRequestArgs
	meEP         *channel.Endpoint
	routeHooks   map[uint64]*routeHook
	routeHooksMu sync.RWMutex
}

func (a *Agent) vips() []net.IP {
	if a.config == nil {
		return nil
	}
	m := make(map[string]struct{})
	m[a.config.Me.String()] = struct{}{}
	for _, o := range a.config.Others {
		m[o.IP.String()] = struct{}{}
	}
	var res []net.IP
	for k := range m {
		ip := net.ParseIP(k)
		if ip != nil {
			res = append(res, ip)
		}
	}
	return res
}

func (a *Agent) configure(args *jsonmsg.ConfigureRequestArgs) error {
	if a.config != nil {
		return errors.New("agent is already configured")
	}
	// TODO: verify that IPs are in 127.0.0.0/8
	me := args.Me.To4()
	if me == nil {
		return errors.Errorf("unexpected IP %s", args.Me)
	}
	meMAC, err := netstackutil.IP2LinkAddress(me)
	if err != nil {
		return err
	}
	meEP := channel.New(chanSz, mtu, meMAC)
	meNICID, err := netstackutil.IP2NICID(me)
	if err != nil {
		return err
	}
	if terr := a.stack.CreateNIC(meNICID, meEP); terr != nil {
		return errors.New(terr.String())
	}
	if terr := a.stack.AddAddress(meNICID, ipv4.ProtocolNumber, tcpip.Address(me)); terr != nil {
		return errors.New(terr.String())
	}

	if terr := a.stack.SetSpoofing(meNICID, true); terr != nil {
		return errors.New(terr.String())
	}

	if terr := a.stack.SetPromiscuousMode(meNICID, true); terr != nil {
		return errors.New(terr.String())
	}

	a.stack.SetRouteTable([]tcpip.Route{
		{
			Destination: header.IPv4EmptySubnet,
			NIC:         meNICID,
		},
	})

	a.meEP = meEP
	a.config = args

	for _, f := range a.config.Forwards {
		if !a.config.Loopback.Disable {
			if err := loopback.GoLocalForward(a.config.Me, f); err != nil {
				return err
			}
		}
		if err := a.goGonetForward(a.config.Me, f); err != nil {
			return err
		}
	}
	for _, o := range a.config.Others {
		if !a.config.Loopback.Disable {
			if err := loopback.GoOther(a.stack, o); err != nil {
				return err
			}
		}
	}

	if err := a.configureDNS(); err != nil {
		return err
	}

	if a.config.HTTP.Listen != "" || a.config.SOCKS.Listen != "" {
		rv, err := resolver.New(a.config.HostnameMap, a.config.Routes, a.vips(), a.stack, a.config.NameServers, a.sender)
		if err != nil {
			return err
		}

		if a.config.HTTP.Listen != "" {
			if err := a.configureHTTP(rv); err != nil {
				return err
			}
		}

		if a.config.SOCKS.Listen != "" {
			if err := a.configureSOCKS(rv); err != nil {
				return err
			}
		}
	}

	if !a.config.StateDir.Disable {
		if err := statedir.Populate(a.config.StateDir.Path, a.config.HostnameMap); err != nil {
			// not a fatal error
			logrus.WithError(err).Warn("failed to create the state directory")
		}
	}

	if a.config.WriteEtcHosts {
		if err := etchosts.Populate("", a.config.HostnameMap, ".bak.norouter"); err != nil {
			// not a fatal error
			logrus.WithError(err).Warn("failed to write /etc/hosts")
		}
	}

	go a.sendL3Routine()
	return nil
}

func (a *Agent) configureDNS() error {
	var dnsSrv *dns.Server
	for _, f := range a.config.NameServers {
		if f.IP.Equal(a.config.Me) {
			if f.Proto != "tcp" {
				return errors.Errorf("expected \"tcp\", got %q as the built-in DNS port", f.Proto)
			}
			logrus.Debugf("dns virtual TCP port=%d", f.Port)
			if dnsSrv != nil {
				return errors.New("duplicated DNS?")
			}
			var err error
			dnsSrv, err = agentdns.New(a.stack, a.config.Me, int(f.Port), a.config.HostnameMap)
			if err != nil {
				return err
			}
		}
		if !a.config.Loopback.Disable {
			if err := loopback.GoOther(a.stack, f.IPPortProto); err != nil {
				return err
			}
		}
	}
	if dnsSrv != nil {
		go func() {
			if e := dnsSrv.ActivateAndServe(); e != nil {
				panic(e)
			}
		}()
	}
	return nil
}

func (a *Agent) configureHTTP(rv *resolver.Resolver) error {
	logrus.Debugf("http listen=%q", a.config.HTTP.Listen)
	l, err := net.Listen("tcp", a.config.HTTP.Listen)
	if err != nil {
		return err
	}
	httpHandler, err := agenthttp.NewHandler(a.stack, rv)
	if err != nil {
		return err
	}
	srv := &http.Server{Handler: httpHandler}
	go srv.Serve(l)
	return nil
}

func (a *Agent) configureSOCKS(rv *resolver.Resolver) error {
	logrus.Debugf("socks listen=%q (supports SOCKS4/4a/5)", a.config.SOCKS.Listen)
	l, err := net.Listen("tcp", a.config.SOCKS.Listen)
	if err != nil {
		return err
	}
	srv, err := agentsocks.NewServer(a.stack, rv)
	if err != nil {
		return err
	}
	go srv.Serve(l)
	return nil
}

func (a *Agent) goGonetForward(me net.IP, f jsonmsg.Forward) error {
	if f.Proto != "tcp" {
		return errors.Errorf("expected proto be \"tcp\", got %q", f.Proto)
	}
	fullAddr := tcpip.FullAddress{
		Addr: tcpip.Address(me),
		Port: f.ListenPort,
	}
	l, err := gonet.ListenTCP(a.stack, fullAddr, ipv4.ProtocolNumber)
	if err != nil {
		return errors.Wrapf(err, "failed to listen on %q", fullAddr)
	}
	go bicopyutil.BicopyAcceptDial(l, f.Proto, fmt.Sprintf("%s:%d", f.ConnectIP, f.ConnectPort), net.Dial)
	return nil
}

func (a *Agent) sendL3Routine() {
	for {
		pi, ok := a.meEP.ReadContext(context.TODO())
		if !ok {
			logrus.Warn("failed to ReadContext")
			continue
		}
		norouterPkt := &stream.Packet{
			Type: stream.TypeL3,
		}
		for _, v := range pi.Pkt.Views() {
			norouterPkt.Payload = append(norouterPkt.Payload, []byte(v)...)
		}
		if err := a.sender.Send(norouterPkt); err != nil {
			logrus.WithError(err).Warn("failed to call sender.Send")
			continue
		}
	}
}

func (a *Agent) onRecvJSON(pkt *stream.Packet) error {
	logrus.Debugf("received JSON: %q", string(pkt.Payload))
	var msg jsonmsg.Message
	if err := json.Unmarshal(pkt.Payload, &msg); err != nil {
		return err
	}
	switch msg.Type {
	case jsonmsg.TypeRequest:
		var req jsonmsg.Request
		if err := json.Unmarshal(msg.Body, &req); err != nil {
			return err
		}
		return a.onRecvRequest(&req)
	default:
		return errors.Errorf("unexpected message type: %q", msg.Type)
	}
}

func (a *Agent) onRecvRequest(req *jsonmsg.Request) error {
	switch req.Op {
	case jsonmsg.OpConfigure:
		var args jsonmsg.ConfigureRequestArgs
		if err := json.Unmarshal(req.Args, &args); err != nil {
			return err
		}
		return a.onRecvConfigureRequest(req, &args)
	default:
		return errors.Errorf("unexpected JSON op: %q", req.Op)
	}
}

func (a *Agent) onRecvConfigureRequest(req *jsonmsg.Request, args *jsonmsg.ConfigureRequestArgs) error {
	if err := a.configure(args); err != nil {
		return err
	}
	data := jsonmsg.ConfigureResultData{
		Features: version.Features,
		Version:  version.Version,
	}
	dataB, err := json.Marshal(data)
	if err != nil {
		return err
	}
	res := jsonmsg.Result{
		RequestID: req.ID,
		Op:        req.Op,
		Error:     nil,
		Data:      dataB,
	}
	resB, err := json.Marshal(res)
	if err != nil {
		return err
	}
	msg := jsonmsg.Message{
		Type: jsonmsg.TypeResult,
		Body: resB,
	}
	msgB, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	resPkt := &stream.Packet{
		Type:    stream.TypeJSON,
		Payload: msgB,
	}
	if err := a.sender.Send(resPkt); err != nil {
		return err
	}
	return nil
}

func (a *Agent) onRecvL3(pkt *stream.Packet) error {
	if a.config == nil {
		return errors.New("received L3 before configuration")
	}
	dstIP := net.IP(pkt.Payload[16:20])
	if dstIP == nil || dstIP.To4() == nil {
		return errors.Errorf("packet does not contain valid dst")
	}
	v := buffer.NewViewFromBytes(pkt.Payload)
	pb := stack.NewPacketBuffer(stack.PacketBufferOptions{
		Data: v.ToVectorisedView(),
	})
	// Routing mode
	if !dstIP.Equal(a.config.Me) {
		// parse.IPv4 and parse.TCP consume PacketBuffer.Data, so we need to create yet another PacketBuffer with same View here :(
		parsed := stack.NewPacketBuffer(stack.PacketBufferOptions{
			Data: v.ToVectorisedView(),
		})
		if !parse.IPv4(parsed) {
			return errors.New("received non-IPv4 packet")
		}
		if !parse.TCP(parsed) {
			return errors.New("received non-TCP packet")
		}
		tcpHdr := header.TCP(parsed.TransportHeader().View())
		if tcpHdr.Flags()&header.TCPFlagSyn != 0 {
			if err := a.prehookRouteOnSYN(parsed); err != nil {
				logrus.WithError(err).Warn("failed to call hookRouteOnSYN")
			}
		}
	}
	a.meEP.InjectInbound(ipv4.ProtocolNumber, pb)
	return nil
}

type routeHook struct {
	l net.Listener
}

func (a *Agent) prehookRouteOnSYN(parsed *stack.PacketBuffer) error {
	ipv4Hdr := header.IPv4(parsed.NetworkHeader().View())
	tcpHdr := header.TCP(parsed.TransportHeader().View())
	dstIP := net.IP(ipv4Hdr.DestinationAddress())
	fullAddr := tcpip.FullAddress{
		Addr: ipv4Hdr.DestinationAddress(),
		Port: tcpHdr.DestinationPort(),
	}
	fullAddrHash := netstackutil.HashFullAddress(fullAddr)
	a.routeHooksMu.RLock()
	_, ok := a.routeHooks[fullAddrHash]
	a.routeHooksMu.RUnlock()
	if ok {
		// logrus.Debugf("routeHooks: found an existing hook for %s:%d, current hooks=%d", dstIP.String(), tcpHdr.DestinationPort(), len(a.routeHooks))
		return nil
	}
	epFunc := func(ep tcpip.Endpoint) error {
		sOpts := ep.SocketOptions()
		sOpts.SetReuseAddress(true)
		sOpts.SetReusePort(true)
		return nil
	}
	l, err := gonetutil.ListenTCPWithEPFunc(a.stack, fullAddr, ipv4.ProtocolNumber, epFunc)
	if err != nil {
		return errors.Wrapf(err, "failed to listen on %q", fullAddr)
	}
	a.routeHooksMu.Lock()
	// logrus.Debugf("routeHooks: installing  hook for %s:%d, current hooks=%d", dstIP.String(), tcpHdr.DestinationPort(), len(a.routeHooks))
	a.routeHooks[fullAddrHash] = &routeHook{
		l: l,
	}
	a.routeHooksMu.Unlock()
	go func() {
		var (
			dialCount   int
			dialCountMu sync.Mutex
		)
		for {
			acceptConn, err := l.Accept()
			if err != nil {
				if strings.Contains(err.Error(), tcpip.ErrInvalidEndpointState.String()) {
					return
				}
				logrus.WithError(err).Error("failed to accept, retrying..")
				continue
			}
			go func() {
				dialCountMu.Lock()
				dialCount++
				//	logrus.Debugf("routeHooks: increased dialCount, new=%d", dialCount)
				dialCountMu.Unlock()
				defer func() {
					dialCountMu.Lock()
					dialCount--
					//		logrus.Debugf("routeHooks: decreased dialCount, new=%d", dialCount)
					zero := dialCount == 0
					dialCountMu.Unlock()
					if zero {
						//			logrus.Debugf("routeHooks: uninstalling hook for %s:%d", net.IP(fullAddr.Addr), fullAddr.Port)
						if err := a.unhookRoute(fullAddrHash); err != nil {
							logrus.Warn(err)
						}
					}
				}()
				defer acceptConn.Close()
				dialConn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", dstIP, tcpHdr.DestinationPort()))
				if err != nil {
					logrus.Warn(err)
					return
				}
				defer dialConn.Close()
				bicopy.Bicopy(acceptConn, dialConn, nil)
			}()
		}
	}()
	return nil
}

func (a *Agent) unhookRoute(fullAddrHash uint64) error {
	var err error
	a.routeHooksMu.Lock()
	if h, ok := a.routeHooks[fullAddrHash]; ok {
		err = h.l.Close()
	}
	delete(a.routeHooks, fullAddrHash)
	a.routeHooksMu.Unlock()
	return err
}

func (a *Agent) Run() error {
	for {
		pkt, err := a.receiver.Recv()
		if err != nil {
			// most error during Recv (e.g. io.EOF) is critical and requires the process to be restarted
			return errors.Wrap(err, "failed to call receiver.Recv")
		}
		switch pkt.Type {
		case stream.TypeJSON:
			if err := a.onRecvJSON(pkt); err != nil {
				logrus.WithError(err).Warn("failed to call onRecvJSON")
			}
		case stream.TypeL3:
			if err := a.onRecvL3(pkt); err != nil {
				logrus.WithError(err).Warn("failed to call onRecvL3")
			}
		default:
			logrus.Warnf("unknown packet type %d", pkt.Type)
		}
	}
}
