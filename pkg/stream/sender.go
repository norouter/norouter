/*
   Copyright (C) Nippon Telegraph and Telephone Corporation.

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
)

// Sender
type Sender struct {
	io.Writer
	sync.Mutex
}

func (sender *Sender) Send(p *Packet) error {
	var buf bytes.Buffer
	// 4 = (sizeof(Type) + sizeof(Padding)) / 8
	metaHdr := uint32(Magic)<<24 | uint32(4+len(p.Payload))
	if err := binary.Write(&buf, binary.BigEndian, metaHdr); err != nil {
		return err
	}
	if err := binary.Write(&buf, binary.BigEndian, p.Type); err != nil {
		return err
	}
	if err := binary.Write(&buf, binary.BigEndian, p.Padding); err != nil {
		return err
	}
	if err := binary.Write(&buf, binary.BigEndian, p.Payload); err != nil {
		return err
	}
	sender.Lock()
	_, err := io.Copy(sender.Writer, &buf)
	sender.Unlock()
	return err
}
