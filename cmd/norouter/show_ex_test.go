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

package main

import (
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/norouter/norouter/pkg/manager/manifest"
	"github.com/norouter/norouter/pkg/manager/manifest/parsed"
)

func TestExampleManifest(t *testing.T) {
	var raw manifest.Manifest
	if err := yaml.UnmarshalWithOptions([]byte(exampleManifest("")), &raw, yaml.Strict()); err != nil {
		t.Fatal(err)
	}
	if _, err := parsed.New(&raw); err != nil {
		t.Fatal(err)
	}
}
