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

package netstackutil

import (
	"net"

	"github.com/pkg/errors"
	"gvisor.dev/gvisor/pkg/tcpip"
)

func IP2NICID(ip net.IP) (tcpip.NICID, error) {
	if len(ip) != 4 {
		return 0, errors.Errorf("unexpected IP %s", ip.String())
	}
	return tcpip.NICID(ip[0]<<24 | ip[1]<<16 | ip[2]<<8 | ip[3]), nil
}

func IP2LinkAddress(ip net.IP) (tcpip.LinkAddress, error) {
	if len(ip) != 4 {
		return "", errors.Errorf("unexpected IP %s", ip.String())
	}
	return tcpip.LinkAddress(append([]byte{0x42, 0x42}, ip...)), nil
}
