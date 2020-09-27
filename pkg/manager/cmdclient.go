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

package manager

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/norouter/norouter/pkg/manager/manifest/parsed"
	"github.com/norouter/norouter/pkg/stream/jsonmsg"
	"github.com/pkg/errors"
)

type CmdClientSet struct {
	ByVIP map[string]*CmdClient
}

func NewCmdClientSet(pm *parsed.ParsedManifest) (*CmdClientSet, error) {
	ccSet := &CmdClientSet{
		ByVIP: make(map[string]*CmdClient),
	}
	for hostname, h := range pm.Hosts {
		client, err := NewCmdClient(hostname, pm)
		if err != nil {
			return nil, err
		}
		ccSet.ByVIP[h.VIP.String()] = client
	}
	return ccSet, nil
}

// NewCmdClient.
func NewCmdClient(hostname string, pm *parsed.ParsedManifest) (*CmdClient, error) {
	h, ok := pm.Hosts[hostname]
	if !ok {
		return nil, errors.Errorf("unexpected hostname %q", hostname)
	}
	var cmd *exec.Cmd
	if len(h.Cmd) != 0 {
		// e.g. ["docker", "exec", "-i", "host1", "--", "norouter"]
		cmd = exec.Command(h.Cmd[0], h.Cmd[1:]...)
	} else {
		if runtime.GOOS == "linux" {
			cmd = exec.Command("/proc/self/exe")
		} else {
			cmd = exec.Command(os.Args[0])
		}
	}
	cmd.Args = append(cmd.Args, "agent", "--automated")
	configRequestArgs := jsonmsg.ConfigureRequestArgs{
		Me: h.VIP,
	}
	for _, p := range h.Ports {
		configRequestArgs.Forwards = append(configRequestArgs.Forwards, *p)
	}
	for _, pub := range pm.PublicHostPorts {
		if pub.IP.Equal(h.VIP) {
			continue
		}
		configRequestArgs.Others = append(configRequestArgs.Others, *pub)
	}
	configRequestArgsB, err := json.Marshal(configRequestArgs)
	if err != nil {
		return nil, err
	}
	req := jsonmsg.Request{
		ID:   GenerateRequestID(),
		Op:   jsonmsg.OpConfigure,
		Args: configRequestArgsB,
	}
	reqB, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	msg := jsonmsg.Message{
		Type: jsonmsg.TypeRequest,
		Body: reqB,
	}
	msgB, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}
	c := &CmdClient{
		Hostname:         hostname,
		VIP:              h.VIP.String(),
		cmd:              cmd,
		configRequestMsg: msgB,
	}
	return c, nil
}

type CmdClient struct {
	Hostname         string
	VIP              string
	cmd              *exec.Cmd
	configRequestMsg json.RawMessage
}

func (c *CmdClient) String() string {
	return fmt.Sprintf("<%s (%s)> %s", c.Hostname, c.VIP, c.cmd.String())
}
