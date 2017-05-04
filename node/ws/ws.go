/*
wsnode provides the ability to call Mongoose OS RPCs via web socket.
*/
package ws

import (
	"errors"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/jpillora/backoff"
	"github.com/neoautomata/mgos-rpc/node"
	"golang.org/x/net/websocket"
)

const (
	maxRecvSize = 1024 // bytes

	retryFactor   = 2
	retryMin      = 250 * time.Millisecond
	retryMax      = 1 * time.Minute
	retryRedial   = true
	retryAttempts = 5
)

var (
	src      = fmt.Sprintf("wsnode-%d", os.Getpid())
	hostname = "loalhost"
)

func init() {
	// Setup the hostname
	hostname = "localhost"
	if h, err := os.Hostname(); err == nil {
		hostname = h
	}
}

// ws represents an RPC connection to a Mongoose OS node via websocket.
type ws struct {
	name string // human-readable name
	addr string // hostname/ip of the sonoff

	sync.Mutex // protects ws
	ws         *websocket.Conn
	wsID       uint16
	backoff    *backoff.Backoff
}

// New creates a new connection to a Mongoose OS node via wesocket.
func New(name, addr string) (node.Node, error) {
	if name == "" {
		return nil, errors.New("A name is required.")
	}

	if addr == "" {
		return nil, errors.New("An address is required.")
	}

	s := &ws{name: name, addr: addr}

	// Setup the backoff controller
	s.backoff = &backoff.Backoff{
		Min:    retryMin,
		Max:    retryMax,
		Factor: 2,
		Jitter: true,
	}

	if err := s.dial(); err != nil {
		return nil, err
	}

	return s, nil
}

func (n *ws) Name() string {
	return n.name
}

func (n *ws) Address() string {
	return n.addr
}

// RPC calls the specified RPC with the provided arguments. Conveserion of
// argument values to float64 is attempted, but strings are used if this fails.
func (n *ws) RPC(method string, argMap map[string]string) (string, error) {
	n.Lock()
	defer n.Unlock()

	args := node.FormatArgs(argMap)
	req := []byte(fmt.Sprintf(
		`{"method":"%s", "args":{%s}, "src":"%s", "id":%d}`,
		method, args, src, n.wsID))
	n.wsID++

	return n.rpc(req)
}

// internal-only no locking.
func (n *ws) rpc(req []byte) (string, error) {
	if err := n.send(req); err != nil {
		return "", err
	}
	resp, err := n.recv()
	if err != nil {
		return "", fmt.Errorf("read from %q failed: %v", n.name, err)
	}
	return string(resp), nil
}

func (n *ws) dial() error {
	if n.ws != nil {
		if err := n.ws.Close(); err != nil {
			log.Printf("WARN: error closing existing web socket: %v", err)
		}
	}

	ws, err := websocket.Dial("ws://"+n.addr+"/rpc", "", "http://"+hostname+"/")
	if err != nil {
		return err
	}

	n.ws = ws
	return nil
}

func (n *ws) send(payload []byte) error {
	n.backoff.Reset()
	defer n.backoff.Reset() // make sure it's left reset
	var err error

	for i := -1; i < retryAttempts; i++ {
		if _, err = n.ws.Write(payload); err != nil {
			err = fmt.Errorf("write to %q failed: %v", n.name, err)

			time.Sleep(n.backoff.Duration())

			if retryRedial {
				if err := n.dial(); err != nil {
					log.Printf("Error dialing new websocket for retry: %v", err)
				}
			}
		} else {
			break
		}
	}

	return err
}

func (n *ws) recv() ([]byte, error) {
	// retry does not make sense for recv. A response will not be re-sent on a new websocket.
	resp := make([]byte, maxRecvSize)
	respLen, err := n.ws.Read(resp)
	if err != nil {
		return nil, fmt.Errorf("read from %q failed: %v", n.name, err)
	}

	return resp[0:respLen], nil
}
