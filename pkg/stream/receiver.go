package stream

import (
	"bytes"
	"encoding/binary"
	"io"
	"net"
	"sync"

	"github.com/sirupsen/logrus"
)

// Receiver
type Receiver struct {
	io.Reader
	sync.Mutex
	DebugDump bool
}

func (receiver *Receiver) Recv() (*Packet, error) {
	var length uint32 // HeaderLength + len(p)
	receiver.Lock()
	if err := binary.Read(receiver.Reader, binary.LittleEndian, &length); err != nil {
		receiver.Unlock()
		return nil, err
	}
	b := make([]byte, length)
	if err := binary.Read(receiver.Reader, binary.LittleEndian, &b); err != nil {
		receiver.Unlock()
		return nil, err
	}
	receiver.Unlock()
	var (
		srcIP4  [4]byte
		srcPort uint16
		dstIP4  [4]byte
		dstPort uint16
		proto   uint16
		flags   uint16
	)
	br := bytes.NewReader(b)
	if err := binary.Read(br, binary.LittleEndian, &srcIP4); err != nil {
		return nil, err
	}
	if err := binary.Read(br, binary.LittleEndian, &srcPort); err != nil {
		return nil, err
	}
	if err := binary.Read(br, binary.LittleEndian, &dstIP4); err != nil {
		return nil, err
	}
	if err := binary.Read(br, binary.LittleEndian, &dstPort); err != nil {
		return nil, err
	}
	if err := binary.Read(br, binary.LittleEndian, &proto); err != nil {
		return nil, err
	}
	if err := binary.Read(br, binary.LittleEndian, &flags); err != nil {
		return nil, err
	}
	pkt := &Packet{
		SrcIP:   net.IP(srcIP4[:]),
		SrcPort: srcPort,
		DstIP:   net.IP(dstIP4[:]),
		DstPort: dstPort,
		Proto:   Proto(proto),
		Flags:   flags,
		Payload: b[HeaderLength:],
	}

	if receiver.DebugDump && logrus.GetLevel() >= logrus.DebugLevel {
		logrus.Debugf("receiver: Received %s:%d->%s:%d (%v) 0b%b: %q",
			pkt.SrcIP.String(), pkt.SrcPort,
			pkt.DstIP.String(), pkt.DstPort,
			pkt.Proto, pkt.Flags, string(pkt.Payload))
	}
	return pkt, nil
}
