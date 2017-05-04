/*
Node contains the interface definition for a Mongoose-OS node
and common utility functions.

LICENSE

   Copyright 2017 neoautomata

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
package node

import (
	"fmt"
	"strconv"
	"strings"
)

type Node interface {
	RPC(method string, args map[string]string) (string, error)
	Name() string
	Address() string
}

func FormatArgs(argMap map[string]string) string {
	argSlice := make([]string, 0, len(argMap))

	for name, val := range argMap {
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			argSlice = append(argSlice, fmt.Sprintf("%q: %f", name, f))
			continue
		}
		argSlice = append(argSlice, fmt.Sprintf("%q: %q", name, val))
	}

	return strings.Join(argSlice, ", ")
}
