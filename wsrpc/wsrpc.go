/*
Make an RPC to a Mongoose OS node via Web Socket

USAGE

An address must be specified via an --address flag OR the first argument, and a
method must be specified via a --method flag OR as the second argument (first if
the address was provided as a flag). Any remaining arguments must be of the form
key=value, and will be parsed and passed as RPC arguments.

EXAMPLES

  $ wsrpc --adress device-id --method RPC.Describe name=RPC.Hello
  $ wsrpc device-id RPC.Describe name=RPC.Hello

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
package main

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/neoautomata/mgos-rpc/node/ws"
)

var (
	address   = flag.String("address", "", "The MQTT broker address and Mongoose OS device ID in URL format. For example: tcp://user:password@mqtt-borker:1883#mgos-device-id")
	method    = flag.String("method", "", "The RPC method to call.")
	printResp = flag.Bool("print_resp", true, "If a response should be expected and printed for the rpc.")
)

func main() {
	flag.Parse()
	args := flag.Args()

	// If the address was not set pull one off remaining args.
	if *address == "" {
		if len(args) > 0 {
			flag.Set("address", args[0])
			args = args[1:] // shift if off
		} else {
			log.Fatalf("An address was not provided either via --address or as an arg.")
		}
	}

	// If the RPC method was not set pull one off remaining args.
	if *method == "" {
		if len(args) > 0 {
			flag.Set("method", args[0])
			args = args[1:] // shift if off
		} else {
			log.Fatalf("A method was not provided either via --method or as an arg.")
		}
	}

	// process any remaing args into RPC args.
	argMap := make(map[string]string, len(args))
	for _, cliArg := range args {
		split := strings.SplitN(cliArg, "=", 2)
		if len(split) != 2 {
			log.Fatalf("RPC arg %q is not formatted as name=value", cliArg)
		}
		argMap[split[0]] = split[1]
	}

	n, err := ws.New(*address, *address)
	if err != nil {
		log.Fatalf("Failed creating ws: %v\n", err)
	}

	resp, err := n.RPC(*method, argMap)
	if err != nil {
		log.Fatalf("RPC %q failed: %v\n", *method, err)
	}

	if *printResp {
		fmt.Println(resp)
	}

}
