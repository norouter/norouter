package stream

import (
	"encoding/binary"
	"hash/fnv"
	"net"

	"github.com/pkg/errors"
)

func HashFields(srcIP net.IP, srcPort uint16, dstIP net.IP, dstPort uint16, proto Proto) uint64 {
	h := fnv.New64()
	if len(srcIP) != 0 {
		srcIP4 := srcIP.To4()
		if srcIP4 == nil {
			panic(errors.Errorf("unsupported IP address %q", srcIP))
		}
		binary.Write(h, binary.LittleEndian, srcIP4)
	} else {
		binary.Write(h, binary.LittleEndian, net.ParseIP("0.0.0.0"))
	}
	binary.Write(h, binary.LittleEndian, srcPort)
	if len(dstIP) != 0 {
		dstIP4 := dstIP.To4()
		if dstIP4 == nil {
			panic(errors.Errorf("unsupported IP address %q", dstIP))
		}
		binary.Write(h, binary.LittleEndian, dstIP4)
	} else {
		binary.Write(h, binary.LittleEndian, net.ParseIP("0.0.0.0"))
	}
	binary.Write(h, binary.LittleEndian, dstPort)
	binary.Write(h, binary.LittleEndian, proto)
	return h.Sum64()
}
