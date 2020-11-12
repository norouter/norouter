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

package version

type Feature = string

const (
	// Features introduced in v0.2.0:
	FeatureLoopback = "loopback" // Listening on multiple loopback IPs such as 127.0.42.101, 127.0.42.102, ...
	FeatureTCP      = "tcp"      // TCP v4 stream
	// Features introduced in v0.4.0:
	FeatureHTTP             = "http"                   // Listening on HTTP for proxy
	FeatureLoopbackDisable  = "loopback.disable"       // Disabling loopback
	FeatureSOCKS            = "socks"                  // Listening a SOCKS proxy (SOCKS4, SOCKS4a, and SOCKS5)
	FeatureHostAliases      = "hostaliases"            // Creating ~/.norouter/agent/hostaliases file
	FeatureHostAliasesXipIO = "hostaliases.\"xip.io\"" // hostaliases using xip.io
	FeatureEtcHosts         = "etchosts"               // Writing /etc/hosts when possible
	// Features introduced in v0.5.0:
	FeatureRoutes = "routes" // Drawing packets into a specific host. Only meaningful for HTTP and SOCKS proxy modes.
	FeatureDNS    = "dns"    // Built-in DNS (10053/tcp)
	// Features introduced in vX.Y.Z:
	// ...
)

var Features = []Feature{FeatureLoopback, FeatureTCP, FeatureHTTP, FeatureLoopbackDisable, FeatureSOCKS, FeatureHostAliases, FeatureEtcHosts, FeatureRoutes, FeatureDNS}
