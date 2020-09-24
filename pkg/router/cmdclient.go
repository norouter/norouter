package router

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	agentconfig "github.com/norouter/norouter/pkg/agent/config"
	"github.com/norouter/norouter/pkg/router/config"
)

type CmdClientSet struct {
	ByVIP map[string]*CmdClient
}

func NewCmdClientSet(cfg *config.Config) (*CmdClientSet, error) {
	var publicHostPorts []string
	for _, h := range cfg.Hosts {
		for _, p := range h.Ports {
			f, err := agentconfig.ParseForward(p)
			if err != nil {
				return nil, err
			}
			publicHostPorts = append(publicHostPorts, fmt.Sprintf("%s:%d/%s", h.VIP, f.ListenPort, f.Proto))
		}
	}
	ccSet := &CmdClientSet{
		ByVIP: make(map[string]*CmdClient),
	}
	for hostname, h := range cfg.Hosts {
		client, err := NewCmdClient(hostname, h, publicHostPorts)
		if err != nil {
			return nil, err
		}
		ccSet.ByVIP[h.VIP] = client
	}
	return ccSet, nil
}

// NewCmdClient.
func NewCmdClient(hostname string, h config.Host, publicHostPorts []string) (*CmdClient, error) {
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
	cmd.Args = append(cmd.Args, "internal", "agent", "--me", h.VIP)
	for _, port := range h.Ports {
		cmd.Args = append(cmd.Args, "--forward", port)
	}
	for _, pub := range publicHostPorts {
		o, err := agentconfig.ParseOther(pub)
		if err != nil {
			return nil, err
		}
		if o.IP.String() == h.VIP {
			continue
		}
		cmd.Args = append(cmd.Args, "--other", pub)
	}
	c := &CmdClient{
		Hostname: hostname,
		VIP:      h.VIP,
		cmd:      cmd,
	}
	return c, nil
}

type CmdClient struct {
	Hostname string
	VIP      string
	cmd      *exec.Cmd
}

func (c *CmdClient) String() string {
	return fmt.Sprintf("<%s (%s)> %s", c.Hostname, c.VIP, c.cmd.String())
}
