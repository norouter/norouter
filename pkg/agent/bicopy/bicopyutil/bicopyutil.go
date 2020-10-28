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

package bicopyutil

import (
	"net"

	"github.com/norouter/norouter/pkg/agent/bicopy"
	"github.com/sirupsen/logrus"
)

type DialFunc = func(string, string) (net.Conn, error)

func BicopyAcceptDial(l net.Listener, dialProto, dialHost string, dialFunc DialFunc) {
	for {
		acceptConn, err := l.Accept()
		if err != nil {
			logrus.WithError(err).Error("failed to accept")
			continue
		}
		go func() {
			defer acceptConn.Close()
			dialConn, err := dialFunc(dialProto, dialHost)
			if err != nil {
				logrus.WithError(err).Errorf("failed to dial to %q (%q)", dialHost, dialProto)
				return
			}
			defer dialConn.Close()
			bicopy.Bicopy(acceptConn, dialConn, nil)
		}()
	}
}
