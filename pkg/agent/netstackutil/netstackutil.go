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

package netstackutil

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"hash/fnv"
	"io"
	"net"

	"gvisor.dev/gvisor/pkg/tcpip"
)

func IP2NICID(ip net.IP) (tcpip.NICID, error) {
	if len(ip) != 4 {
		return 0, fmt.Errorf("unexpected IP %s", ip.String())
	}
	return tcpip.NICID(ip[0]<<24 | ip[1]<<16 | ip[2]<<8 | ip[3]), nil
}

func IP2LinkAddress(ip net.IP) (tcpip.LinkAddress, error) {
	if len(ip) != 4 {
		return "", fmt.Errorf("unexpected IP %s", ip.String())
	}
	return tcpip.LinkAddress(append([]byte{0x42, 0x42}, ip...)), nil
}

func HashFullAddress(fa tcpip.FullAddress) uint64 {
	var buf bytes.Buffer
	if err := binary.Write(&buf, binary.BigEndian, fa.NIC); err != nil {
		panic(err)
	}
	if _, err := buf.Write([]byte(fa.Addr)); err != nil {
		panic(err)
	}
	if err := binary.Write(&buf, binary.BigEndian, fa.Port); err != nil {
		panic(err)
	}
	h := fnv.New64a()
	if _, err := io.Copy(h, &buf); err != nil {
		panic(err)
	}
	return h.Sum64()
}
