/*
mqtt provides the ability to call Mongoose OS RPCs via web socket.

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
package mqtt

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	client "github.com/eclipse/paho.mqtt.golang"
	"github.com/neoautomata/mgos-rpc/node"
)

const (
	srcBase = "mqttNode"
)

// mqttNode represents an RPC connection to a Mongoose OS node via MQTT.
type mqttNode struct {
	sync.Mutex

	name     string // human-readable name
	deviceID string
	conn     client.Client
	id       int
	rchan    chan []byte
	src      string
}

// New creates a new connection to a Mongoose OS node via wesocket.
func New(name, deviceID string, conn client.Client) (node.Node, error) {
	if name == "" {
		return nil, errors.New("name is required")
	}

	if deviceID == "" {
		return nil, errors.New("deviceID is required")
	}

	if conn == nil {
		return nil, errors.New("an MQTT connection is required")
	}

	n := &mqttNode{name: name, deviceID: deviceID, conn: conn, rchan: make(chan []byte)}

	// generate a unique name for this instance.
	b := make([]byte, 4) // 4 bytes == 8 hex chars
	if _, err := rand.Read(b); err != nil {
		return nil, fmt.Errorf("error reading random bytes: %v", err)
	}
	n.src = fmt.Sprintf("%s-%x", srcBase, b)

	// Subscribe to receive responses.
	st := conn.Subscribe(fmt.Sprintf("%s/rpc", n.src), 2 /* Exactly Once */, n.recv)
	if ok := st.WaitTimeout(30 * time.Second); !ok {
		return nil, errors.New("Timeout waiting for subscription.")
	}
	if st.Error() != nil {
		return nil, st.Error()
	}

	return n, nil
}

func (n *mqttNode) recv(c client.Client, msg client.Message) {
	type resp struct {
		ID       int
		Src, Dst string
	}

	r := new(resp)
	if err := json.Unmarshal(msg.Payload(), r); err != nil {
		log.Printf("WARN: failed parsing MQTT message: %v", err)
	}
	if r.Dst == n.src && r.Src == n.deviceID && r.ID == n.id-1 {
		n.rchan <- msg.Payload()
	} else {
		log.Printf("WARN: Ignoring MQTT message: %s", string(msg.Payload()))
	}
}

func (n *mqttNode) Name() string {
	return n.name
}

func (n *mqttNode) Address() string {
	return n.deviceID
}

// RPC calls the specified RPC with the provided arguments. Conveserion of
// argument values to float64 is attempted, but strings are used if this fails.
func (n *mqttNode) RPC(method string, argMap map[string]string) (string, error) {
	n.Lock()
	defer n.Unlock()

	args := node.FormatArgs(argMap)
	req := []byte(fmt.Sprintf(
		`{"method":"%s", "args":{%s}, "src":"%s", "id":%d}`,
		method, args, n.src, n.id))
	n.id++

	pt := n.conn.Publish(fmt.Sprintf("%s/rpc", n.deviceID), 2 /* Eactly Once */, false /* not retained */, req)
	if ok := pt.WaitTimeout(30 * time.Second); !ok {
		return "", errors.New("Timeout waiting for publish")
	}
	if pt.Error() != nil {
		return "", fmt.Errorf("publish to %q failed: %v", n.name, pt.Error())
	}

	return string(<-n.rchan), nil
}
