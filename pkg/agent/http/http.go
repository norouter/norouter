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

package http

import (
	"bufio"
	"context"
	"net"
	"net/http"
	"net/url"
	"strconv"

	"github.com/elazarl/goproxy"
	"github.com/norouter/norouter/pkg/bicopy"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/adapters/gonet"
	"gvisor.dev/gvisor/pkg/tcpip/network/ipv4"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
)

// NewHandlerHandler returns a http.Handler that works as proxy.
func NewHandler(st *stack.Stack, hostnameMap map[string]net.IP) (http.Handler, error) {
	p := goproxy.NewProxyHttpServer()
	for xhostname, xip := range hostnameMap {
		hostname := xhostname
		ip := xip
		var cond goproxy.ReqConditionFunc = func(req *http.Request, ctx *goproxy.ProxyCtx) bool {
			s := req.URL.Hostname()
			return s == hostname || s == ip.String()
		}
		var doFunc = func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
			resp, err := do(st, ip, req, ctx)
			if err != nil {
				logrus.WithError(err).Warn("failed to call do()")
				return req, goproxy.NewResponse(req,
					goproxy.ContentTypeText, http.StatusInternalServerError,
					"See NoRouter agent log\n")
			}
			return req, resp
		}
		p.OnRequest(cond).DoFunc(doFunc)
		var hijackFunc = func(req *http.Request, clientConn net.Conn, ctx *goproxy.ProxyCtx) {
			defer clientConn.Close()
			if err := hijack(st, ip, req, clientConn, ctx); err != nil {
				logrus.WithError(err).Warn("failed to call hijack()")
				clientConn.Write([]byte("HTTP/1.1 500 Cannot reach destination\r\n\r\n"))
			}
		}
		p.OnRequest(cond).HijackConnect(hijackFunc)
	}
	return p, nil
}

func do(st *stack.Stack, ip net.IP, req *http.Request, ctx *goproxy.ProxyCtx) (*http.Response, error) {
	gonetDialConn, err := gonetDial(st, ip, req)
	if err != nil {
		return nil, err
	}
	// Do NOT defer close gonetDialConn
	remoteBuf := bufio.NewReadWriter(bufio.NewReader(gonetDialConn), bufio.NewWriter(gonetDialConn))
	if err := req.Write(gonetDialConn); err != nil {
		return nil, err
	}
	if err := remoteBuf.Flush(); err != nil {
		return nil, err
	}
	resp, err := http.ReadResponse(remoteBuf.Reader, req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func hijack(st *stack.Stack, ip net.IP, req *http.Request, clientConn net.Conn, ctx *goproxy.ProxyCtx) error {
	gonetDialConn, err := gonetDial(st, ip, req)
	if err != nil {
		return err
	}
	defer gonetDialConn.Close()
	clientConn.Write([]byte("HTTP/1.1 200 Connection established\r\n\r\n"))
	bicopy.Bicopy(clientConn, gonetDialConn, nil)
	return nil
}

func gonetDial(st *stack.Stack, ip net.IP, req *http.Request) (net.Conn, error) {
	port, err := portNumFromURL(req.URL)
	if err != nil {
		return nil, err
	}
	fullAddr := tcpip.FullAddress{
		Addr: tcpip.Address(ip),
		Port: uint16(port),
	}
	return gonet.DialContextTCP(context.TODO(), st, fullAddr, ipv4.ProtocolNumber)
}

func portNumFromURL(u *url.URL) (int, error) {
	s := u.Port()
	if s != "" {
		return strconv.Atoi(s)
	}
	switch u.Scheme {
	case "http":
		return 80, nil
	case "https":
		return 443, nil
	}
	return 0, errors.Errorf("url seems to lack port: %q", u.String())
}
