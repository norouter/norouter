package router

import (
	"os"

	"github.com/norouter/norouter/pkg/stream"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

func New(ccSet *CmdClientSet) (*Router, error) {
	r := &Router{
		ccSet:     ccSet,
		senders:   make(map[string]*stream.Sender),
		receivers: make(map[string]*stream.Receiver),
	}
	return r, nil
}

type Router struct {
	ccSet     *CmdClientSet
	senders   map[string]*stream.Sender // key: vip (TODO: don't use string)
	receivers map[string]*stream.Receiver
}

func (r *Router) Run() error {
	debugDump := false
	// Step 1: fill up senders
	for vip, cc := range r.ccSet.ByVIP {
		cc.cmd.Stderr = &stderrWriter{
			vip:      cc.VIP,
			hostname: cc.Hostname,
		}
		stdin, err := cc.cmd.StdinPipe()
		if err != nil {
			return err
		}
		sender := &stream.Sender{
			Writer:    stdin,
			DebugDump: debugDump,
		}
		r.senders[vip] = sender
		stdout, err := cc.cmd.StdoutPipe()
		if err != nil {
			return err
		}
		receiver := &stream.Receiver{
			Reader:    stdout,
			DebugDump: debugDump,
		}
		r.receivers[vip] = receiver
		logrus.Infof("starting client for %s(%s): %q", cc.Hostname, vip, cc.cmd.String())
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
				dstIPStr := pkt.DstIP.String()
				sender, ok := r.senders[dstIPStr]
				if !ok {
					logrus.WithError(err).Warnf("unexpected dstIP %s in a packet from %s", dstIPStr, vip)
					continue
				}
				logrus.Debugf("routing packet from %s:%d to %s:%d", pkt.SrcIP, pkt.SrcPort, pkt.DstIP, pkt.DstPort)
				if err := sender.Send(pkt); err != nil {
					logrus.WithError(err).Warnf("routing packet from %s:%d to %s:%d", pkt.SrcIP, pkt.SrcPort, pkt.DstIP, pkt.DstPort)
					continue
				}
			}
		})
	}
	return eg.Wait()
}

type stderrWriter struct {
	hostname string
	vip      string
}

func (w *stderrWriter) Write(p []byte) (int, error) {
	logrus.Warnf("stderr[%s(%s)]: %s", w.hostname, w.vip, string(p))
	return len(p), nil
}
