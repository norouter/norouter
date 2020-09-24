package agent

import (
	"encoding/binary"
	"fmt"
	"hash/fnv"
	"io"
	"math/rand"
	"net"
	"runtime"
	"sync"
	"time"

	"github.com/norouter/norouter/pkg/agent/config"
	"github.com/norouter/norouter/pkg/agent/conn"
	"github.com/norouter/norouter/pkg/bicopy"
	"github.com/norouter/norouter/pkg/debugutil"
	"github.com/norouter/norouter/pkg/stream"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func runForward(me net.IP, f *config.Forward) error {
	lh := fmt.Sprintf("%s:%d", me.String(), f.ListenPort)
	l, err := net.Listen(f.Proto, lh)
	if err != nil {
		return errors.Wrapf(err, "failed to listen on %q", lh)
	}
	go func() {
		for {
			lconn, err := l.Accept()
			if err != nil {
				logrus.WithError(err).Error("failed to accept")
				continue
			}
			go func() {
				dh := fmt.Sprintf("%s:%d", f.ConnectIP.String(), f.ConnectPort)
				dconn, err := net.Dial("tcp", dh)
				if err != nil {
					logrus.WithError(err).Errorf("failed to dial to %q", dh)
					return
				}
				defer dconn.Close()
				defer lconn.Close()
				bicopy.Bicopy(lconn, dconn, nil)
			}()
		}
	}()
	return nil
}

func newRandSource(me net.IP, now time.Time) rand.Source {
	h := fnv.New64()
	me4 := me.To4()
	if me4 == nil {
		panic(errors.Errorf("unsupported IP address %q", me))

	}
	binary.Write(h, binary.LittleEndian, me)
	binary.Write(h, binary.LittleEndian, now.UnixNano)
	seed := h.Sum64()
	return rand.NewSource(int64(seed))
}

func New(me net.IP, others []*config.Other, forwards []*config.Forward, w io.Writer, r io.Reader) (*Agent, error) {
	debugDump := false
	sender := &stream.Sender{
		Writer:    w,
		DebugDump: debugDump,
	}
	receiver := &stream.Receiver{
		Reader:    r,
		DebugDump: debugDump,
	}
	a := &Agent{
		me:          me,
		others:      others,
		tcpForwards: make(map[uint16]*config.Forward),
		sender:      sender,
		receiver:    receiver,
		rand:        rand.New(newRandSource(me, time.Now())),
		pwHM:        make(map[uint64]*io.PipeWriter),
	}
	for _, f := range forwards {
		if f.Proto != "tcp" {
			return nil, errors.Errorf("unexpected proto %q", f.Proto)
		}
		a.tcpForwards[f.ListenPort] = f
	}
	return a, nil
}

type Agent struct {
	me          net.IP
	others      []*config.Other
	tcpForwards map[uint16]*config.Forward // key: listenPort
	sender      *stream.Sender
	receiver    *stream.Receiver
	rand        *rand.Rand
	pwHM        map[uint64]*io.PipeWriter
	pwHMMu      sync.RWMutex
}

func (a *Agent) generateVSrcPort() uint16 {
	var vSrcPort uint16
	for {
		vSrcPort = uint16(a.rand.Int())
		if vSrcPort == 0 {
			continue
		}
		if _, conflict := a.tcpForwards[vSrcPort]; conflict {
			continue
		}
		// FIXME: detect more colisions
		break
	}
	return vSrcPort
}

func (a *Agent) runOther(o *config.Other) error {
	lh := fmt.Sprintf("%s:%d", o.IP, o.Port)
	l, err := net.Listen(o.Proto, lh)
	if err != nil {
		return err
	}
	go func() {
		for {
			lconn, err := l.Accept()
			if err != nil {
				logrus.WithError(err).Error("failed to accept")
				continue
			}
			vSrcPort := a.generateVSrcPort()
			logrus.Debugf("generated vSrcPort=%d", vSrcPort)
			conn, pw, err := conn.New(a.me, vSrcPort, o.IP, o.Port, a.sender)
			if err != nil {
				logrus.WithError(err).Warn("failed to create conn")
				lconn.Close()
				return
			}
			hdrHash := stream.HashFields(o.IP, o.Port, a.me, vSrcPort, stream.TCP)
			a.pwHMMu.Lock()
			logrus.Debugf("registering pw for %s:%d->%s:%d", o.IP, o.Port, a.me, vSrcPort)
			a.pwHM[hdrHash] = pw
			a.pwHMMu.Unlock()
			go func() {
				defer lconn.Close()
				defer conn.Close()
				bicopy.Bicopy(conn, lconn, nil)
			}()
		}
	}()
	return nil
}

func (a *Agent) getPW(hdrHash uint64, pkt *stream.Packet) (*io.PipeWriter, bool, error) {
	a.pwHMMu.RLock()
	pw, pwOk := a.pwHM[hdrHash]
	a.pwHMMu.RUnlock()
	if pwOk {
		return pw, pwOk, nil
	}
	if f, fOk := a.tcpForwards[pkt.DstPort]; fOk {
		// connect to forward ports, e.g. 8080 (->127.0.0.1:80)
		dh := fmt.Sprintf("%s:%d", f.ConnectIP.String(), f.ConnectPort)
		dconn, err := net.Dial(f.Proto, dh)
		if err != nil {
			logrus.WithError(err).Warnf("failed to dial %q", dh)
			return nil, false, err
		}
		logrus.Debugf("dialed to %q, creating replyConn", dh)
		var replyConn *conn.Conn
		replyConn, pw, err = conn.New(pkt.DstIP, pkt.DstPort, pkt.SrcIP, pkt.SrcPort, a.sender)
		if err != nil {
			dconn.Close()
			return nil, false, errors.Wrap(err, "failed to create replyConn")
		}
		a.pwHMMu.Lock()
		logrus.Debugf("registering pw for %s:%d->%s:%d", pkt.SrcIP, pkt.SrcPort, pkt.DstIP, pkt.DstPort)
		a.pwHM[hdrHash] = pw
		a.pwHMMu.Unlock()
		go func() {
			defer dconn.Close()
			defer replyConn.Close()
			bicopy.Bicopy(dconn, replyConn, nil)
		}()
		return pw, true, nil
	}
	return nil, false, nil
}

func (a *Agent) Run() error {
	for _, f := range a.tcpForwards {
		if err := runForward(a.me, f); err != nil {
			return err
		}
	}
	for _, o := range a.others {
		if err := a.runOther(o); err != nil {
			return err
		}
	}
	for {
		pkt, err := a.receiver.Recv()
		if err != nil {
			return errors.Wrap(err, "failed to recv from receiver")
		}
		if pkt.Proto != stream.TCP {
			logrus.Warnf("received unknown proto %d, ignoring", pkt.Proto)
			continue
		}
		if !pkt.DstIP.Equal(a.me) {
			logrus.Warnf("received dstIP=%s is not me (%s),  ignoring", pkt.DstIP.String(), a.me.String())
			continue
		}
		hdrHash := stream.HashFields(pkt.SrcIP, pkt.SrcPort, pkt.DstIP, pkt.DstPort, pkt.Proto)
		pw, pwOk, err := a.getPW(hdrHash, pkt)
		if err != nil {
			logrus.WithError(err).Warnf("failed to call getPW (%s:%d->%s:%d)", pkt.SrcIP, pkt.SrcPort, pkt.DstIP, pkt.DstPort)
		}
		if pwOk {
			logrus.Debugf("Calling pw.Write %s:%d->%s:%d", pkt.SrcIP, pkt.SrcPort, pkt.DstIP, pkt.DstPort)
			if _, err := pw.Write(pkt.Payload); err != nil {
				logrus.WithError(err).Warn("pw.Write failed")
			}
		} else {
			logrus.Debugf("NOT calling pw.Write %s:%d->%s:%d", pkt.SrcIP, pkt.SrcPort, pkt.DstIP, pkt.DstPort)
		}
		a.gc(pkt, hdrHash, pw)
		a.debugPrintStat()
	}
}

func (a *Agent) gc(pkt *stream.Packet, hdrHash uint64, pw *io.PipeWriter) {
	// FIXME: support half-closing properly
	if pkt.Flags&stream.FlagCloseRead != 0 || pkt.Flags&stream.FlagCloseWrite != 0 {
		if pw != nil {
			if err := pw.Close(); err != nil {
				logrus.WithError(err).Debugf("failed to close pw")
			}
		}
		a.pwHMMu.Lock()
		delete(a.pwHM, hdrHash)
		a.pwHMMu.Unlock()
	}
}

func (a *Agent) debugPrintStat() {
	if logrus.GetLevel() >= logrus.DebugLevel {
		a.pwHMMu.RLock()
		l := len(a.pwHM)
		a.pwHMMu.RUnlock()
		logrus.Debugf("STAT: len(a.pwHM)=%d,GoRoutines=%d, FDs=%d", l, runtime.NumGoroutine(), debugutil.NumFDs())
	}
}
