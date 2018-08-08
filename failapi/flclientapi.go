package failapi

import (
	"github.com/pebbe/zmq4"
	"log"
	"time"

	"fmt"
	"strconv"
)

const (
	GlobalTimeout = 3000 * time.Millisecond
	PingInterval  = 2000 * time.Millisecond
	ServerTTL     = 6000 * time.Millisecond
)

type FailoverClientApi struct {
	pipe *zmq4.Socket
}

// Simple structure  for one server required for backend agent
type Server struct {
	endpoint string
	alive    bool
	pingAt   time.Time
	expires  time.Time
}

//Message processing agent
type Agent struct {
	pipe     *zmq4.Socket
	router   *zmq4.Socket
	servers  map[string]*Server
	actives  []*Server
	sequence int
	reply    []string
	request  []string
	expires  time.Time
}

func New() (flcapi *FailoverClientApi) {
	flcapi = &FailoverClientApi{}

	var err error
	flcapi.pipe, err = zmq4.NewSocket(zmq4.PAIR)
	if err != nil {
		log.Fatalln(err)
	}

	flcapi.pipe.Bind("inproc://pipe")
	go failoverClientApiAgent()
	return
}

//  The front-end object sends a multi-part message to the back-end agent.
func (flcapi *FailoverClientApi) Connect(endpoint string) {
	flcapi.pipe.SendMessage("CONNECT", endpoint)

	//  Allow connection to come up
	time.Sleep(100 * time.Millisecond)
}

// The front-end object sends a message to the back-end
func (flcapi *FailoverClientApi) Request(request []string) (reply []string, err error) {
	flcapi.pipe.SendMessage("REQUEST", request)
	reply, err = flcapi.pipe.RecvMessage(0)
	if err == nil {
		status := reply[0]
		reply = reply[1:]
		if status == "FAILED" {
			reply = reply[0:0]
		}
	}

	return
}

func NewServer(endpoint string) (srv *Server) {
	srv = &Server{
		endpoint: endpoint,
		alive:    false,
		pingAt:   time.Now().Add(PingInterval),
		expires:  time.Now().Add(ServerTTL),
	}

	return
}

func (srv *Server) ping(socket *zmq4.Socket) {
	if time.Now().After(srv.pingAt) {
		socket.SendMessage(srv.endpoint, "PING")
		srv.pingAt = time.Now().Add(PingInterval)
	}
}

func (srv *Server) tickless(t time.Time) time.Time {
	if t.After(srv.pingAt) {
		return srv.pingAt
	}

	return t
}

func NewAgent() (agent *Agent) {
	agent = &Agent{
		servers: make(map[string]*Server),
		actives: make([]*Server, 0),
		request: make([]string, 0),
		reply:   make([]string, 0),
	}

	var err error
	agent.pipe, err = zmq4.NewSocket(zmq4.PAIR)
	if err != nil {
		log.Fatalln(err)
	}

	agent.pipe.Connect("inproc://pipe")
	agent.router, err = zmq4.NewSocket(zmq4.ROUTER)
	if err != nil {
		log.Fatalln(err)
	}

	return
}

// Control message processor
func (agent *Agent) controlMessage() {
	msg, err := agent.pipe.RecvMessage(0)
	if err != nil {
		log.Fatalln(err)
	}

	command := msg[0]
	msg = msg[1:]

	switch command {
	case "CONNECT":
		endpoint := msg[0]
		fmt.Printf("Connecting to %s...\n", endpoint)
		err := agent.router.Connect(endpoint)
		if err != nil {
			log.Fatalln(err)
		}

		server := NewServer(endpoint)
		agent.servers[endpoint] = server
		agent.actives = append(agent.actives, server)
		server.pingAt = time.Now().Add(PingInterval)
		server.expires = time.Now().Add(ServerTTL)

	case "REQUEST":
		if len(agent.request) > 0 {
			log.Fatalln("len(agent.request) > 0")
		}

		agent.request = make([]string, 1, 1+len(msg))
		agent.sequence++
		agent.request[0] = fmt.Sprint(agent.sequence)
		agent.request = append(agent.request, msg...)
		agent.expires = time.Now().Add(GlobalTimeout)
	}
}

//Router message processor
func (agent *Agent) routerMessage() {
	reply, err := agent.router.RecvMessage(0)
	if err != nil {
		log.Fatalln(err)
	}

	endpoint := reply[0]
	reply = reply[1:]
	server, ok := agent.servers[endpoint]
	if !ok {
		log.Fatalln("No server endpoint")
	}

	if !server.alive {
		agent.actives = append(agent.actives, server)
		server.alive = true
	}

	server.pingAt = time.Now().Add(PingInterval)
	server.expires = time.Now().Add(ServerTTL)

	sequence, _ := strconv.Atoi(reply[0])

	reply = reply[1:]
	if sequence == agent.sequence {
		agent.pipe.SendMessage("OK", reply)
		agent.request = agent.request[0:0]
	}
}

func failoverClientApiAgent() {
	agent := NewAgent()

	poller := zmq4.NewPoller()
	poller.Add(agent.pipe, zmq4.POLLIN)
	poller.Add(agent.router, zmq4.POLLIN)

	for {
		tickless := time.Now().Add(time.Hour)
		if len(agent.request) > 0 && tickless.After(agent.expires) {
			tickless = agent.expires
		}

		for key := range agent.servers {
			tickless = agent.servers[key].tickless(tickless)
		}

		polled, err := poller.Poll(tickless.Sub(time.Now()))
		if err != nil {
			break
		}

		for _, item := range polled {
			switch item.Socket {
			case agent.pipe:
				agent.controlMessage()
			case agent.router:
				agent.routerMessage()
			}
		}

		//if request is processed dispatch to next server
		if len(agent.request) > 0 {
			if time.Now().After(agent.expires) {
				agent.pipe.SendMessage("FAILED")
				agent.request = agent.request[0:0]
			} else {
				// find server, remove expired one
				for len(agent.actives) > 0 {
					server := agent.actives[0]
					if time.Now().After(server.expires) {
						agent.actives = agent.actives[1:]
						server.alive = false
					} else {
						agent.router.SendMessage(server.endpoint, agent.request)
						break
					}
				}
			}

		}

		for key := range agent.servers {
			agent.servers[key].ping(agent.router)
		}
	}
}
