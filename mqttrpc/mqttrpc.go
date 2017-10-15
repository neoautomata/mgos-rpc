/*
Make an RPC to a Mongoose OS node via MQTT

USAGE

An address must be specified via an --address flag OR the first argument, and a
method must be specified via a --method flag OR as the second argument (first if
the address was provided as a flag). Any remaining arguments must be of the form
key=value, and will be parsed and passed as RPC arguments.

EXAMPLES

  $ mqttrpc --adress tcp://user:pass@broker:1883#device-id --method RPC.Describe name=RPC.Hello
  $ mqttrpc tcp://user:pass@broker:1883#device-id RPC.Describe name=RPC.Hello

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
	"net/url"
	"strings"
	"time"

	client "github.com/eclipse/paho.mqtt.golang"
	"github.com/neoautomata/mgos-rpc/node/mqtt"
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

	u, err := url.Parse(*address)
	if err != nil {
		log.Fatalf("Failed parsing MQTT URL: %v", err)
	}
	if u.Fragment == "" {
		log.Fatalf("Address %q does not include a device id as a URL fragment", *address)
	}

	co := client.NewClientOptions()
	co.AddBroker(fmt.Sprintf("%s://%s", u.Scheme, u.Host))

	if u.User != nil {
		co.SetUsername(u.User.Username())
		if pass, ok := u.User.Password(); ok {
			co.SetPassword(pass)
		}
	}

	co.SetConnectTimeout(30 * time.Second)
	mqttConn := client.NewClient(co)

	ct := mqttConn.Connect()
	if ok := ct.WaitTimeout(30 * time.Second); !ok {
		log.Fatal("Timed out waiting for MQTT connection")
	}
	if ct.Error() != nil {
		log.Fatal(ct.Error())
	}

	n, err := mqtt.New(u.Fragment, u.Fragment, mqttConn)
	if err != nil {
		log.Fatalf("Failed creating MQTT node: %v", err)
	}

	resp, err := n.RPC(*method, argMap)
	if err != nil {
		log.Fatalf("RPC %q failed: %v\n", *method, err)
	}

	if *printResp {
		fmt.Println(resp)
	}

}
