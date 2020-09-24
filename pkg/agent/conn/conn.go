package conn

import (
	"io"
	"net"

	"github.com/hashicorp/go-multierror"
	"github.com/norouter/norouter/pkg/stream"
)

func New(srcIP net.IP, srcPort uint16, dstIP net.IP, dstPort uint16, sender *stream.Sender) (*Conn, *io.PipeWriter, error) {
	pr, pw := io.Pipe()
	c := &Conn{
		SrcIP:   srcIP,
		SrcPort: srcPort,
		DstIP:   dstIP,
		DstPort: dstPort,
		sender:  sender,
		pr:      pr,
	}
	return c, pw, nil
}

type Conn struct {
	SrcIP   net.IP
	SrcPort uint16
	DstIP   net.IP
	DstPort uint16

	sender      *stream.Sender
	pr          *io.PipeReader // pkt.Payload, passed from receiver
	readClosed  bool
	writeClosed bool
}

func (c *Conn) Write(p []byte) (int, error) {
	pkt := &stream.Packet{
		SrcIP:   c.SrcIP,
		SrcPort: c.SrcPort,
		DstIP:   c.DstIP,
		DstPort: c.DstPort,
		Proto:   stream.TCP,
		Flags:   0,
		Payload: p,
	}
	if err := c.sender.Send(pkt); err != nil {
		return 0, err
	}
	return len(p), nil
}

func (c *Conn) Read(p []byte) (int, error) {
	return c.pr.Read(p)
}

func (c *Conn) Close() error {
	if c.readClosed && c.writeClosed {
		return nil
	}
	if c.readClosed {
		return c.CloseWrite()
	}
	if c.writeClosed {
		return c.CloseRead()
	}
	var merr *multierror.Error
	pkt := &stream.Packet{
		SrcIP:   c.SrcIP,
		SrcPort: c.SrcPort,
		DstIP:   c.DstIP,
		DstPort: c.DstPort,
		Proto:   stream.TCP,
		Flags:   stream.FlagCloseRead | stream.FlagCloseWrite,
		Payload: nil,
	}
	if err := c.sender.Send(pkt); err != nil {
		merr = multierror.Append(merr, err)
	}
	if err := c.pr.Close(); err != nil {
		merr = multierror.Append(merr, err)
	}
	c.readClosed = true
	c.writeClosed = true
	return merr.ErrorOrNil()
}

func (c *Conn) CloseRead() error {
	if c.readClosed {
		return nil
	}
	var merr *multierror.Error
	pkt := &stream.Packet{
		SrcIP:   c.SrcIP,
		SrcPort: c.SrcPort,
		DstIP:   c.DstIP,
		DstPort: c.DstPort,
		Proto:   stream.TCP,
		Flags:   stream.FlagCloseRead,
		Payload: nil,
	}
	if err := c.sender.Send(pkt); err != nil {
		merr = multierror.Append(merr, err)
	}
	if err := c.pr.Close(); err != nil {
		merr = multierror.Append(merr, err)
	}
	c.readClosed = true
	return merr.ErrorOrNil()
}

func (c *Conn) CloseWrite() error {
	if c.writeClosed {
		return nil
	}
	var merr *multierror.Error
	pkt := &stream.Packet{
		SrcIP:   c.SrcIP,
		SrcPort: c.SrcPort,
		DstIP:   c.DstIP,
		DstPort: c.DstPort,
		Proto:   stream.TCP,
		Flags:   stream.FlagCloseWrite,
		Payload: nil,
	}
	if err := c.sender.Send(pkt); err != nil {
		merr = multierror.Append(merr, err)
	}
	c.writeClosed = true
	return merr.ErrorOrNil()
}
