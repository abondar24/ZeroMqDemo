package rpsapi

import (
	"fmt"
	"github.com/pebbe/zmq4"
	"log"
	"strconv"
	"time"
)

var pipeNum uint64

const (
	serverMax    = 2
	serverTTL    = 5000 * time.Millisecond
	stateINITIAL = iota
	stateSYNCING //get state from server
	stateACTIVE  //get new updates from server
)

type RelPubSubClient struct {
	pipe *zmq4.Socket
}

//backend agent
type backendServer struct {
	address    string
	port       int
	snapshot   *zmq4.Socket
	subscriber *zmq4.Socket
	expiry     time.Time
	requests   int64
}

type agent struct {
	pipe       *zmq4.Socket
	kvmap      map[string]*KVmsg
	subtree    string
	server     [serverMax]*backendServer
	serversNum int
	state      int
	curServer  int
	sequence   int64
	publisher  *zmq4.Socket
}

func NewClient() (client *RelPubSubClient) {
	client = &RelPubSubClient{}
	var err error
	client.pipe, err = zmq4.NewSocket(zmq4.PAIR)
	if err != nil {
		log.Fatalln(err)
	}

	pipeName := fmt.Sprint("inproc://pipe", pipeNum)
	pipeNum++
	client.pipe.Bind(pipeName)
	go rpsAgent(pipeName)
	return
}

func newBackendServer(address string, port int, subtree string) (server *backendServer) {
	server = &backendServer{}

	fmt.Printf("Adding server %s:%d...\n", address, port)
	server.address = address
	server.port = port

	var err error

	server.snapshot, err = zmq4.NewSocket(zmq4.DEALER)
	if err != nil {
		log.Fatalln(err)
	}
	server.snapshot.Connect(fmt.Sprintf("%s:%d", address, port))

	server.subscriber, err = zmq4.NewSocket(zmq4.SUB)
	if err != nil {
		log.Fatalln(err)
	}
	server.subscriber.Connect(fmt.Sprintf("%s:%d", address, port+1))
	server.subscriber.SetSubscribe(subtree)

	return
}

func newAgent(pipe *zmq4.Socket) (agt *agent) {
	agt = &agent{}
	agt.pipe = pipe
	agt.kvmap = make(map[string]*KVmsg)
	agt.subtree = ""
	agt.state = stateINITIAL

	var err error
	agt.publisher, err = zmq4.NewSocket(zmq4.PUB)
	if err != nil {
		log.Fatalln(err)
	}

	return
}

func (client *RelPubSubClient) Subtree(subtree string) {
	client.pipe.SendMessage("SUBTREE", subtree)
}

func (client *RelPubSubClient) Connect(address, service string) {
	client.pipe.SendMessage("CONNECT", address, service)
}

//  Set a new value in the shared hashmap.
func (client *RelPubSubClient) Set(key, value string, ttl int) {
	client.pipe.SendMessage("SET", key, value, ttl)
}

// Lookup a value in shared hashmap
func (client *RelPubSubClient) Get(key string) (value string, err error) {
	client.pipe.SendMessage("GET", key)

	reply, err := client.pipe.RecvMessage(0)
	if err != nil {
		return
	}

	value = reply[0]
	return
}

func (agt *agent) controlMessageHandler() (err error) {
	msg, err := agt.pipe.RecvMessage(0)
	if err != nil {
		return
	}

	cmd := msg[0]
	msg = msg[1:]

	switch cmd {
	case "SUBTREE":
		agt.subtree = msg[0]

	case "CONNECT":
		address := msg[0]
		service := msg[1]

		if agt.serversNum < serverMax {
			srv, _ := strconv.Atoi(service)
			agt.server[agt.serversNum] = newBackendServer(address, srv, agt.subtree)
			agt.serversNum++
			agt.publisher.Connect(fmt.Sprintf("%s:%d", address, srv+2))
		} else {
			fmt.Printf("Too many servers(max, %d)\n", serverMax)
		}

	case "SET":
		key := msg[0]
		val := msg[1]
		ttl := msg[2]

		kvmsg := NewKVMessage(0)
		kvmsg.SetKey(key)
		kvmsg.SetUUID()
		kvmsg.SetBody(val)
		kvmsg.SetProp("ttl", ttl)
		kvmsg.StoreMsg(agt.kvmap)
		kvmsg.SendKVmsg(agt.publisher)

	case "GET":
		key := msg[0]
		value := ""

		if kvmsg, ok := agt.kvmap[key]; ok {
			value, _ = kvmsg.GetBody()
		}

		agt.pipe.SendMessage(value)
	}

	return
}

//async agent which manages a server pool and handles the req/rep dialog
func rpsAgent(pipename string) {
	pipe, err := zmq4.NewSocket(zmq4.PAIR)
	if err != nil {
		log.Fatalln(err)
	}

	pipe.Connect(pipename)
	agt := newAgent(pipe)

	for {
		poller := zmq4.NewPoller()
		poller.Add(pipe, zmq4.POLLIN)
		server := agt.server[agt.curServer]

		switch agt.state {
		case stateINITIAL:
			//ask for a snapshot
			if agt.serversNum > 0 {
				fmt.Printf("Waiting for server at %s:%d ", server.address, server.port)
				if server.requests < 2 {
					server.snapshot.SendMessage("ICANHAZ?", agt.subtree)
					server.requests++
				}

				server.expiry = time.Now().Add(serverTTL)
				agt.state = stateSYNCING
				poller.Add(server.snapshot, zmq4.POLLIN)
			}
		case stateSYNCING:
			//read from snapshot
			poller.Add(server.snapshot, zmq4.POLLIN)
		case stateACTIVE:
			//read from subscriber
			poller.Add(server.subscriber, zmq4.POLLIN)
			break
		}

		pollTimer := time.Duration(-1)
		if server != nil {
			pollTimer = server.expiry.Sub(time.Now())
			if pollTimer < 0 {
				pollTimer = 0
			}
		}

		polled, err := poller.Poll(pollTimer)
		if err != nil {
			break
		}

		if len(polled) > 0 {
			for _, item := range polled {
				switch socket := item.Socket; socket {
				case pipe:
					err = agt.controlMessageHandler()
					if err != nil {
						log.Fatalln(err)
					}
				default:
					kvmsg, err := RecvKVmsg(socket)
					if err != nil {
						log.Fatalln(err)
					}

					//reset expriry if we get something
					server.expiry = time.Now().Add(serverTTL)
					if agt.state == stateSYNCING {
						server.requests = 0
						if key, _ := kvmsg.GetKey(); key == "KTHXBAI" {
							agt.sequence, _ = kvmsg.GetSequence()
							agt.state = stateACTIVE
							fmt.Printf("Received from %s:%d snapshot=%d\n", server.address, server.port, agt.sequence)
						} else {
							kvmsg.StoreMsg(agt.kvmap)
						}
					} else if agt.state == stateACTIVE {
						if seq, _ := kvmsg.GetSequence(); seq > agt.sequence {
							agt.sequence = seq
							kvmsg.StoreMsg(agt.kvmap)
							fmt.Printf("Received from %s:%d update=%d\n", server.address, server.port, agt.sequence)
						}
					}
				}
			}
		} else {
			fmt.Printf("Server at %s:%d didn't give HUGZ\n", server.address, server.port)
			agt.curServer = (agt.curServer + 1) % agt.serversNum
			agt.state = stateINITIAL
		}
	}
}
