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

package jsonmsg

import "encoding/json"

type Type = string

const (
	TypeRequest = "request" // Main to sub
	TypeResult  = "result"  // Sub to main
	// Other possible types: "debug_log", "notify", ...
)

type Message struct {
	Type Type            `json:"type"` // Required
	Body json.RawMessage `json:"body"` // Request or Result
}

type Op = string

type Request struct {
	ID   int             `json:"id"` // Required
	Op   Op              `json:"op"` // Required
	Args json.RawMessage `json:"args,omitempty"`
}

type Result struct {
	RequestID int             `json:"request_id"` // Required
	Op        Op              `json:"op"`         // Required
	Error     json.RawMessage `json:"error,omitempty"`
	Data      json.RawMessage `json:"data,omitempty"`
}
