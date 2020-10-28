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

	agenthttp "github.com/norouter/norouter/pkg/agent/http"
	"github.com/norouter/norouter/pkg/agent/loopback"
	"github.com/norouter/norouter/pkg/bicopy/bicopyutil"
	"github.com/norouter/norouter/pkg/netstackutil"
	"github.com/norouter/norouter/pkg/stream"
	"github.com/norouter/norouter/pkg/stream/jsonmsg"
	"github.com/norouter/norouter/pkg/version"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/adapters/gonet"
	"gvisor.dev/gvisor/pkg/tcpip/buffer"
	"gvisor.dev/gvisor/pkg/tcpip/header"
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
		HandleLocal:        true,
	}
	st := stack.New(opts)
	st.SetForwarding(ipv4.ProtocolNumber, true)
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
	stack    *stack.Stack
	sender   *stream.Sender
	receiver *stream.Receiver
	config   *jsonmsg.ConfigureRequestArgs
	meEP     *channel.Endpoint
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
	if a.config.HTTP.Listen != "" {
		logrus.Debugf("http listen=%q", a.config.HTTP.Listen)
		l, err := net.Listen("tcp", a.config.HTTP.Listen)
		if err != nil {
			return err
		}
		httpHandler, err := agenthttp.NewHandler(a.stack, a.config.HostnameMap)
		if err != nil {
			return err
		}
		srv := &http.Server{Handler: httpHandler}
		go srv.Serve(l)
	}
	go a.sendL3Routine()
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
	go bicopyutil.BicopyAcceptDial(l, f.Proto, fmt.Sprintf("%s:%d", f.ConnectIP.String(), f.ConnectPort), net.Dial)
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
	pb := stack.NewPacketBuffer(stack.PacketBufferOptions{
		Data: buffer.NewViewFromBytes(pkt.Payload).ToVectorisedView(),
	})
	a.meEP.InjectInbound(ipv4.ProtocolNumber, pb)
	return nil
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
