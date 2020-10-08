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

package stream

import (
	"bytes"
	"encoding/binary"
	"io"
	"sync"

	"github.com/pkg/errors"
)

// Receiver
type Receiver struct {
	io.Reader
	sync.Mutex
}

func (receiver *Receiver) Recv() (*Packet, error) {
	var metaHdr uint32
	receiver.Lock()
	if err := binary.Read(receiver.Reader, binary.BigEndian, &metaHdr); err != nil {
		receiver.Unlock()
		return nil, err
	}
	if magic := uint8(metaHdr >> 24); magic != Magic {
		receiver.Unlock()
		return nil, errors.Errorf("expected magic to be 0x%x, got 0x%x", Magic, magic)
	}
	length := metaHdr & 0xFFFFFF
	b := make([]byte, length)
	if err := binary.Read(receiver.Reader, binary.BigEndian, &b); err != nil {
		receiver.Unlock()
		return nil, err
	}
	receiver.Unlock()
	br := bytes.NewReader(b)
	var (
		typ     Type
		padding uint16
	)
	if err := binary.Read(br, binary.BigEndian, &typ); err != nil {
		return nil, err
	}
	if err := binary.Read(br, binary.BigEndian, &padding); err != nil {
		return nil, err
	}
	pkt := &Packet{
		Type:    typ,
		Padding: padding,
		Payload: b[4:],
	}
	return pkt, nil
}
