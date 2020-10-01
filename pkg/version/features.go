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

package version

type Feature = string

const (
	// Features introduced in v0.2.0:
	FeatureLoopback = "loopback" // Listening on loopback IPs
	FeatureTCP      = "tcp"      // TCP v4 stream
	// Features introduced in vX.Y.Z:
	// ...
)

var Features = []Feature{FeatureLoopback, FeatureTCP}
