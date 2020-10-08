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

package manager

import (
	"encoding/json"
	"net"
	"os"

	"github.com/norouter/norouter/pkg/stream"
	"github.com/norouter/norouter/pkg/stream/jsonmsg"
	"github.com/norouter/norouter/pkg/version"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

func New(ccSet *CmdClientSet) (*Manager, error) {
	r := &Manager{
		ccSet:     ccSet,
		senders:   make(map[string]*stream.Sender),
		receivers: make(map[string]*stream.Receiver),
	}
	return r, nil
}

type Manager struct {
	ccSet     *CmdClientSet
	senders   map[string]*stream.Sender // key: vip (TODO: don't use string)
	receivers map[string]*stream.Receiver
}

func (r *Manager) Run() error {
	// Step 1: fill up senders
	for vip, cc := range r.ccSet.ByVIP {
		configPkt := &stream.Packet{
			Type:    stream.TypeJSON,
			Payload: cc.configRequestMsg,
		}
		cc.cmd.Stderr = &stderrWriter{
			vip:      cc.VIP,
			hostname: cc.Hostname,
		}
		stdin, err := cc.cmd.StdinPipe()
		if err != nil {
			return err
		}
		sender := &stream.Sender{
			Writer: stdin,
		}
		r.senders[vip] = sender
		stdout, err := cc.cmd.StdoutPipe()
		if err != nil {
			return err
		}
		receiver := &stream.Receiver{
			Reader: stdout,
		}
		r.receivers[vip] = receiver
		logrus.Debugf("starting client for %s (%s): %q", cc.Hostname, vip, cc.cmd.String())
		if err := cc.cmd.Start(); err != nil {
			return err
		}
		// TODO: notify if a client exits
		defer func() {
			logrus.Warnf("exiting client: %q", cc.String())
			if err := cc.cmd.Process.Signal(os.Interrupt); err != nil {
				logrus.WithError(err).Errorf("error while sending os.Interrupt to %s(%s)", cc.Hostname, vip)
				cc.cmd.Process.Kill()
			}
		}()
		logrus.Debugf("sending Configure packet to %s: %q", cc.Hostname, string(cc.configRequestMsg))
		if err := sender.Send(configPkt); err != nil {
			return err
		}
	}

	var eg errgroup.Group
	// Step 2: start goroutines after filling up all r.senders
	for vipx, receiverx := range r.receivers {
		vip := vipx
		receiver := receiverx
		eg.Go(func() error {
			for {
				pkt, err := receiver.Recv()
				if err != nil {
					return errors.Errorf("failed to receive from %s", vip)
				}
				switch pkt.Type {
				case stream.TypeJSON:
					if err := r.onRecvJSON(vip, pkt); err != nil {
						logrus.WithError(err).Warn("error while handling JSON packet")
					}
				case stream.TypeL3:
					if err := r.onRecvL3(vip, pkt); err != nil {
						logrus.WithError(err).Warn("error while handling L3 packet")
					}
				default:
					logrus.WithError(err).Warnf("unexpected packet type %d", pkt.Type)
				}
			}
		})
	}
	return eg.Wait()
}

func (r *Manager) onRecvJSON(vip string, pkt *stream.Packet) error {
	var msg jsonmsg.Message
	if err := json.Unmarshal(pkt.Payload, &msg); err != nil {
		return err
	}
	switch msg.Type {
	case jsonmsg.TypeResult:
		var res jsonmsg.Result
		if err := json.Unmarshal(msg.Body, &res); err != nil {
			return err
		}
		return r.onRecvResult(vip, &res)
	default:
		return errors.Errorf("unexpected JSON message type: %q", msg.Type)
	}
}

func (r *Manager) onRecvResult(vip string, res *jsonmsg.Result) error {
	if len(res.Error) != 0 {
		return errors.Errorf("got an error result %q", res.Error)
	}
	switch res.Op {
	case jsonmsg.OpConfigure:
		var data jsonmsg.ConfigureResultData
		if err := json.Unmarshal(res.Data, &data); err != nil {
			return err
		}
		return r.onRecvConfigureResult(vip, data)
	default:
		return errors.Errorf("unexpected JSON op: %q", res.Op)
	}
}

func (r *Manager) onRecvConfigureResult(vip string, data jsonmsg.ConfigureResultData) error {
	logrus.Debugf("received ConfigureResult from %s: %+v", vip, data)
	if data.Version != version.Version {
		logrus.Warnf("version mismatch on %s: %s vs %s", vip, data.Version, version.Version)
	}
	fm := make(map[string]struct{})
	for _, f := range data.Features {
		fm[f] = struct{}{}
	}
	if _, ok := fm[version.FeatureTCP]; !ok {
		return errors.Errorf("%s lacks essential feature %q", vip, version.FeatureTCP)
	}
	logrus.Infof("Ready: %s", vip)
	return nil
}

func (r *Manager) onRecvL3(vip string, pkt *stream.Packet) error {
	dstIP := net.IP(pkt.Payload[16:20])
	if dstIP == nil || dstIP.To4() == nil {
		return errors.Errorf("packet does not contain valid dst")
	}
	dstIPStr := dstIP.To4().String()
	sender, ok := r.senders[dstIPStr]
	if !ok {
		return errors.Errorf("unexpected dstIP %s in a packet from %s", dstIPStr, vip)
	}
	if err := sender.Send(pkt); err != nil {
		return err
	}
	return nil
}

type stderrWriter struct {
	hostname string
	vip      string
}

func (w *stderrWriter) Write(p []byte) (int, error) {
	logrus.Warnf("stderr[%s(%s)]: %s", w.hostname, w.vip, string(p))
	return len(p), nil
}
