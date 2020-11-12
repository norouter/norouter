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

package router

import (
	"net"
	"testing"

	"github.com/norouter/norouter/pkg/stream/jsonmsg"
	"gotest.tools/v3/assert"
)

func TestRouter(t *testing.T) {
	routes := []jsonmsg.Route{
		{
			ToCIDR: []string{"192.168.95.0/24", "192.168.96.0/24"},
			Via:    net.ParseIP("127.0.42.101"),
		},
		{
			ToCIDR: []string{"192.168.96.0/24", "192.168.97.0/24"},
			Via:    net.ParseIP("127.0.42.102"),
		},
		{
			ToCIDR: []string{"192.168.96.200/32"},
			Via:    net.ParseIP("127.0.42.101"),
		},
	}
	testCases := map[string]string{
		"192.168.95.1":   "127.0.42.101",
		"192.168.95.2":   "127.0.42.101",
		"192.168.96.1":   "127.0.42.102",
		"192.168.97.1":   "127.0.42.102",
		"192.168.98.1":   "192.168.98.1",
		"192.168.96.200": "127.0.42.101",
	}
	r, err := New(routes)
	assert.NilError(t, err)
	for to, expected := range testCases {
		assert.Equal(t, expected, r.Route(net.ParseIP(to)).String())
	}
	r.Learn([]net.IP{net.ParseIP("192.168.95.1")}, net.ParseIP("127.0.42.150"), true)
	assert.Equal(t, "127.0.42.150", r.Route(net.ParseIP("192.168.95.1")).String())
}

func TestRouterNil(t *testing.T) {
	testCases := map[string]string{
		"127.0.42.101": "127.0.42.101",
		"192.168.98.1": "192.168.98.1",
	}
	r, err := New(nil)
	assert.NilError(t, err)
	for to, expected := range testCases {
		assert.Equal(t, expected, r.Route(net.ParseIP(to)).String())
	}
}

func TestRouterZero(t *testing.T) {
	routes := []jsonmsg.Route{
		{
			ToCIDR: []string{"0.0.0.0/0"},
			Via:    net.ParseIP("127.0.42.101"),
		},
	}
	testCases := map[string]string{
		"192.168.98.1": "127.0.42.101",
	}
	r, err := New(routes)
	assert.NilError(t, err)
	for to, expected := range testCases {
		assert.Equal(t, expected, r.Route(net.ParseIP(to)).String())
	}
}

func TestRouterZeroPlus(t *testing.T) {
	routes := []jsonmsg.Route{
		{
			ToCIDR: []string{"0.0.0.0/0"},
			Via:    net.ParseIP("127.0.42.101"),
		},
		{
			ToCIDR: []string{"192.168.99.0/24"},
			Via:    net.ParseIP("127.0.42.102"),
		},
	}
	testCases := map[string]string{
		"192.168.98.1": "127.0.42.101",
		"192.168.99.1": "127.0.42.102",
	}
	r, err := New(routes)
	assert.NilError(t, err)
	for to, expected := range testCases {
		assert.Equal(t, expected, r.Route(net.ParseIP(to)).String())
	}
}

func TestRouterHostname(t *testing.T) {
	routes := []jsonmsg.Route{
		{
			ToHostnameGlob: []string{"*.cloud1.example.com", "*.cloud2.example.com"},
			Via:            net.ParseIP("127.0.42.101"),
		},
		{
			ToHostnameGlob: []string{"*.cloud2.example.com", "*.cloud3.example.com"},
			Via:            net.ParseIP("127.0.42.102"),
		},
		{
			ToHostnameGlob: []string{"foo.cloud3.example.com"},
			Via:            net.ParseIP("127.0.42.101"),
		},
		{
			ToHostnameGlob: []string{"bar.cloud3.example.com."}, // canonical
			Via:            net.ParseIP("127.0.42.103"),
		},
	}
	testCases := map[string]string{
		"host1.cloud1.example.com":  "127.0.42.101",
		"host1.cloud1.example.com.": "127.0.42.101", // canonical
		"host2.cloud1.example.com":  "127.0.42.101",
		"host1.cloud2.example.com":  "127.0.42.102",
		"host1.cloud3.example.com":  "127.0.42.102",
		"foo.cloud3.example.com":    "127.0.42.101",
		"bar.cloud3.example.com":    "127.0.42.103",
		"none.example.com":          "<nil>",
	}
	r, err := New(routes)
	assert.NilError(t, err)
	for to, expected := range testCases {
		assert.Equal(t, expected, r.RouteWithHostname(to).String())
	}
}
