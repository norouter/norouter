package stream

import (
	"net"
)

type Proto = uint16

const (
	TCP Proto = 0
)

// HeaderLength:
//
// + uint64 srcIP   (4 bytes)
//
// + uint16 srcPort (2 bytes)
//
// + uint32 dstIP   (4 bytes)
//
// + uint16 dstPort (2 bytes)
//
// + uint16 proto   (2 bytes)
//
// + uint32 flags   (2 bytes)
const HeaderLength = 4 + 2 + 4 + 2 + 2 + 2

// Packet requires uint32le length to be prepended.
// The protocol is highly likely to be changed.
type Packet struct {
	// SrcIP is the src IP. Must be [4]byte.
	SrcIP net.IP
	// SrcPort is the dest port.
	SrcPort uint16
	// DstIP is the dest IP. Must be [4]byte.
	DstIP net.IP
	// DstPort is the dest port.
	DstPort uint16
	// Proto must be TCP.
	Proto Proto
	// Flags, such as FlagCloseRead and FlagCloseWrite
	Flags uint16
	// Payload does not contain any L2/L3/L4 headers.
	Payload []byte
}

const (
	FlagCloseRead  = 0b01
	FlagCloseWrite = 0b10
)
