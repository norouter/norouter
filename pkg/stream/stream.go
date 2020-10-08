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

type Type = uint16

const (
	Magic            = uint8(0x42)
	TypeInvalid Type = 0x0
	TypeL3      Type = 0x1
	TypeJSON    Type = 0x2
)

// Packet requires uint32be length to be prepended.
// The upper 8 bits of the length must be Magic
type Packet struct {
	Type    Type
	Padding uint16
	Payload []byte // L3 or JSON
}
