package config

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type Other struct {
	IP    net.IP
	Port  uint16
	Proto string
}

func (o *Other) String() string {
	return fmt.Sprintf("%s:%d/%s", o.IP.String(), o.Port, o.Proto)
}

type Forward struct {
	// listenIP is "me"
	ListenPort  uint16
	ConnectIP   net.IP
	ConnectPort uint16
	Proto       string
}

// parseMe parses --me=127.0.42.101 flag
func ParseMe(me string) (net.IP, error) {
	ip := net.ParseIP(me)
	if ip == nil {
		return nil, errors.Errorf("invalid \"me\" IP %q", me)
	}
	ip = ip.To4()
	if ip == nil {
		return nil, errors.Errorf("invalid \"me\" IP %q, must be IPv4", me)
	}
	return ip, nil
}

// ParseOther parses --other=127.0.42.102:8080[/tcp] flag
func ParseOther(other string) (*Other, error) {
	s := strings.TrimSuffix(other, "/tcp")
	if strings.Contains(s, "/") {
		// TODO: support "/udp" suffix
		return nil, errors.Errorf("cannot parse \"other\" address %q", other)
	}
	h, p, err := net.SplitHostPort(s)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot parse \"other\" address %q", other)
	}
	ip := net.ParseIP(h)
	if ip == nil {
		return nil, errors.Errorf("cannot parse \"other\" address %q", other)
	}
	port, err := strconv.Atoi(p)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot parse \"other\" address %q", other)
	}
	o := &Other{
		IP:    ip,
		Port:  uint16(port),
		Proto: "tcp",
	}
	return o, nil
}

// ParseForward parses --forward=8080:127.0.0.1:80[/tcp] flag
func ParseForward(forward string) (*Forward, error) {
	s := strings.TrimSuffix(forward, "/tcp")
	if strings.Contains(s, "/") {
		// TODO: support "/udp" suffix
		return nil, errors.Errorf("cannot parse \"forward\" address %q", forward)
	}
	split := strings.Split(s, ":")
	if len(split) != 3 {
		return nil, errors.Errorf("cannot parse \"forward\" address %q", forward)
	}
	listenPort, err := strconv.Atoi(split[0])
	if err != nil {
		return nil, errors.Wrapf(err, "cannot parse \"forward\" address %q", forward)
	}
	connectIP := net.ParseIP(split[1])
	if connectIP == nil {
		return nil, errors.Errorf("cannot parse \"forward\" address %q", forward)
	}
	connectPort, err := strconv.Atoi(split[2])
	if err != nil {
		return nil, errors.Wrapf(err, "cannot parse \"forward\" address %q", forward)
	}
	f := &Forward{
		ListenPort:  uint16(listenPort),
		ConnectIP:   connectIP,
		ConnectPort: uint16(connectPort),
		Proto:       "tcp",
	}
	return f, nil
}
