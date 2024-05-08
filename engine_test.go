// Copyright 2024 justlorain
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ezcache

import (
	"fmt"
	"testing"
)

func TestRun(t *testing.T) {
	addrs := []string{"http://localhost:7246", "http://localhost:7247", "http://localhost:7248"}
	e := NewEngine(WithAddr(addrs[0]))
	e.RegisterNodes(addrs...)
	if err := e.Run(); err != nil {
		fmt.Println(err)
	}
}
