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

package builtinports

const (
	// DNSTCP is the default TCP port number of the built-in DNS.
	// The port number was chosen so that it can be associated with a loopback device without the root privileges.
	// Note that resolv.conf does not support specifying non-53 port.
	DNSTCP = 10053
)
