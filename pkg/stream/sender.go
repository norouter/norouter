package stream

import (
	"bytes"
	"encoding/binary"
	"io"
	"sync"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// Sender
type Sender struct {
	io.Writer
	sync.Mutex
	DebugDump bool
}

func (sender *Sender) Send(p *Packet) error {
	if sender.DebugDump && logrus.GetLevel() >= logrus.DebugLevel {
		logrus.Debugf("sender: Sending %s:%d %s:%d (%v) 0b%b: %q",
			p.SrcIP.String(), p.SrcPort,
			p.DstIP.String(), p.DstPort,
			p.Proto, p.Flags, string(p.Payload))
	}
	var buf bytes.Buffer
	if err := binary.Write(&buf, binary.LittleEndian, uint32(HeaderLength+len(p.Payload))); err != nil {
		return err
	}
	sip := p.SrcIP.To4()
	if sip == nil {
		return errors.Errorf("unexpected ip %+v", sip)
	}
	if err := binary.Write(&buf, binary.LittleEndian, []byte(sip)); err != nil {
		return err
	}
	if err := binary.Write(&buf, binary.LittleEndian, p.SrcPort); err != nil {
		return err
	}
	dip := p.DstIP.To4()
	if dip == nil {
		return errors.Errorf("unexpected ip %+v", dip)
	}
	if err := binary.Write(&buf, binary.LittleEndian, []byte(dip)); err != nil {
		return err
	}
	if err := binary.Write(&buf, binary.LittleEndian, p.DstPort); err != nil {
		return err
	}
	if err := binary.Write(&buf, binary.LittleEndian, uint16(p.Proto)); err != nil {
		return err
	}
	if err := binary.Write(&buf, binary.LittleEndian, p.Flags); err != nil {
		return err
	}
	if p.Payload != nil {
		if err := binary.Write(&buf, binary.LittleEndian, p.Payload); err != nil {
			return err
		}
	}
	sender.Lock()
	_, err := io.Copy(sender.Writer, &buf)
	sender.Unlock()
	return err
}
