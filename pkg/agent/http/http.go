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
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"

	"github.com/elazarl/goproxy"
	"github.com/hashicorp/go-multierror"
	"github.com/norouter/norouter/pkg/agent/bicopy"
	"github.com/norouter/norouter/pkg/agent/resolver"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/adapters/gonet"
	"gvisor.dev/gvisor/pkg/tcpip/network/ipv4"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
)

// NewHandlerHandler returns a http.Handler that works as proxy.
func NewHandler(st *stack.Stack, rv *resolver.Resolver) (http.Handler, error) {
	p := goproxy.NewProxyHttpServer()
	var cond goproxy.ReqConditionFunc = func(req *http.Request, ctx *goproxy.ProxyCtx) bool {
		s := req.URL.Hostname()
		return rv.Interesting(s)
	}
	var doFunc = func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		resp, err := do(st, rv, req, ctx)
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
		if err := hijack(st, rv, req, clientConn, ctx); err != nil {
			logrus.WithError(err).Warn("failed to call hijack()")
			clientConn.Write([]byte("HTTP/1.1 500 Cannot reach destination\r\n\r\n"))
		}
	}
	p.OnRequest(cond).HijackConnect(hijackFunc)
	return p, nil
}

func do(st *stack.Stack, rv *resolver.Resolver, req *http.Request, ctx *goproxy.ProxyCtx) (*http.Response, error) {
	gonetDialConn, err := gonetDial(st, rv, req)
	if err != nil {
		return nil, err
	}
	// Do NOT defer close gonetDialConn
	remoteBuf := bufio.NewReadWriter(bufio.NewReader(gonetDialConn), bufio.NewWriter(gonetDialConn))
	if err := req.Write(gonetDialConn); err != nil {
		gonetDialConn.Close()
		return nil, err
	}
	if err := remoteBuf.Flush(); err != nil {
		gonetDialConn.Close()
		return nil, err
	}
	resp, err := http.ReadResponse(remoteBuf.Reader, req)
	if err != nil {
		gonetDialConn.Close()
		return nil, err
	}
	rc := &readCloserWithHook{
		Reader:    resp.Body,
		klose:     resp.Body.Close,
		postClose: gonetDialConn.Close,
	}
	resp.Body = rc
	return resp, nil
}

type readCloserWithHook struct {
	io.Reader
	klose     func() error
	postClose func() error
}

func (rc *readCloserWithHook) Close() error {
	var merr *multierror.Error
	if err := rc.klose(); err != nil {
		merr = multierror.Append(merr, err)
	}
	if rc.postClose != nil {
		if err := rc.postClose(); err != nil {
			merr = multierror.Append(merr, err)
		}
	}
	return merr.ErrorOrNil()
}

func hijack(st *stack.Stack, rv *resolver.Resolver, req *http.Request, clientConn net.Conn, ctx *goproxy.ProxyCtx) error {
	gonetDialConn, err := gonetDial(st, rv, req)
	if err != nil {
		return err
	}
	defer gonetDialConn.Close()
	clientConn.Write([]byte("HTTP/1.1 200 Connection established\r\n\r\n"))
	bicopy.Bicopy(clientConn, gonetDialConn, nil)
	return nil
}

func gonetDial(st *stack.Stack, rv *resolver.Resolver, req *http.Request) (net.Conn, error) {
	ip, err := rv.Resolve(req.URL.Hostname())
	if err != nil {
		return nil, err
	}
	port, err := portNumFromURL(req.URL)
	if err != nil {
		return nil, err
	}
	fullAddr := tcpip.FullAddress{
		Addr: tcpip.Address(ip),
		Port: uint16(port),
	}
	conn, err := gonet.DialContextTCP(context.TODO(), st, fullAddr, ipv4.ProtocolNumber)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to dial gonet %s:%d", ip, port)
	}
	return conn, nil
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
