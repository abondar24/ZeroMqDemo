package haServerApi

import (
	"errors"
	"github.com/pebbe/zmq4"
	"log"
	"strconv"
	"time"
)

const (
	Primary = true
	Backup  = false
)

type state int

const (
	_ = state(iota)
	statePrimary
	stateBackup
	stateActive
	statePassive
)

type event int

const (
	_ = event(iota)
	peerPrimary
	peerBackup
	peerActive
	peerPassive
	clientRequest
)

type HAServer struct {
	Reactor    *zmq4.Reactor
	statepub   *zmq4.Socket
	statesub   *zmq4.Socket
	state      state
	event      event
	peerExpiry time.Time
	voter      func(socket *zmq4.Socket) error
	active     func() error
	passive    func() error
}

const Heartbeat = 1000 * time.Millisecond

func (srv *HAServer) executeStateMachine() (err error) {
	if srv.state == statePrimary {
		if srv.event == peerBackup {
			log.Println("connected to backup(passive) server. ready as active")
			srv.state = stateActive
			if srv.active != nil {
				srv.active()
			}
		} else if srv.event == peerActive {
			log.Println("connected to backup(active) server. ready as passive")
			srv.state = statePassive
			if srv.passive != nil {
				srv.passive()
			}
		} else if srv.event == clientRequest {
			if time.Now().After(srv.peerExpiry) {
				log.Println("request from client. ready as active")
				srv.state = stateActive
				if srv.active != nil {
					srv.active()
				}
			} else {
				err = errors.New("performing a failback and the backup is currently active")
			}
		}
	} else if srv.state == stateBackup {
		if srv.event == peerActive {
			log.Println("connected to primary(active) server. ready as passive")
			srv.state = statePassive
			if srv.passive != nil {
				srv.passive()
			}
		} else if srv.event == clientRequest {
			err = errors.New("error")
		}
	} else if srv.state == stateActive {
		if srv.event == peerActive {
			log.Println("both servers are in active state - aborting")
			err = errors.New("both servers are in active state")
		}
	} else if srv.state == statePassive {
		if srv.event == peerPrimary {
			log.Println("primary(passive) is restarting, ready as active")
			srv.state = stateActive
		} else if srv.event == peerBackup {
			log.Println("backup(passive) restarting. ready as active")
			srv.state = stateActive
		} else if srv.event == peerPassive {
			log.Println("both servers are in passive state - aborting")
			err = errors.New("both servers are in passive state")
		} else if srv.event == clientRequest {
			if time.Now().After(srv.peerExpiry) {
				log.Println("failover successful. ready as active")
				srv.state = stateActive
			} else {
				err = errors.New("peer is alive, reject connections")
			}
		}

		if srv.state == stateActive && srv.active != nil {
			srv.active()
		}
	}

	return
}

func (srv *HAServer) updatePeerExpiry() {
	srv.peerExpiry = time.Now().Add(2 * Heartbeat)
}

func (srv *HAServer) sendState() (err error) {
	_, err = srv.statepub.SendMessage(int(srv.state))
	return
}

func (srv *HAServer) recvState() (err error) {
	msg, err := srv.statesub.RecvMessage(0)
	if err == nil {
		evt, err := strconv.Atoi(msg[0])
		if err != nil {
			return err
		}

		srv.event = event(evt)
	}

	return srv.executeStateMachine()
}

func (srv *HAServer) voterReady(socket *zmq4.Socket) error {
	srv.event = clientRequest
	err := srv.executeStateMachine()
	if err == nil {
		srv.voter(socket)
	} else {
		socket.RecvMessage(0)
	}

	return nil
}

func NewServer(primary bool, local, remote string) (srv *HAServer, err error) {

	srv = &HAServer{}

	srv.Reactor = zmq4.NewReactor()

	if primary {
		srv.state = statePrimary
	} else {
		srv.state = stateBackup
	}

	srv.statepub, err = zmq4.NewSocket(zmq4.PUB)
	if err != nil {
		log.Println(err)
		return &HAServer{}, err
	}
	srv.statepub.Bind(local)

	srv.statesub, err = zmq4.NewSocket(zmq4.SUB)
	if err != nil {
		log.Println(err)
		return &HAServer{}, err
	}

	srv.statesub.SetSubscribe("")
	srv.statesub.Connect(remote)

	srv.Reactor.AddChannelTime(time.Tick(Heartbeat), 1,
		func(i interface{}) error {
			return srv.sendState()
		})

	srv.Reactor.AddSocket(srv.statesub, zmq4.POLLIN,
		func(e zmq4.State) error {
			return srv.recvState()
		})

	return
}

//  The voter method registers a client voter socket. Messages received
//  on this socket provide the clientRequest events for the server
//  FSM and are passed to the provided application handler.
//  exactly one voter per server instance is required

func (srv *HAServer) Voter(endpoint string, socketType zmq4.Type, handler func(socket *zmq4.Socket) error) {
	socket, err := zmq4.NewSocket(socketType)
	if err != nil {
		log.Fatalln(err)
	}

	socket.Bind(endpoint)
	if srv.voter != nil {
		log.Fatalln("Double voter function")
	}

	srv.voter = handler
	srv.Reactor.AddSocket(socket, zmq4.POLLIN, func(evt zmq4.State) error {
		return srv.voterReady(socket)
	})
}

func (srv *HAServer) RegisterActiveHandler(handler func() error) {
	if srv.active != nil {
		log.Fatalln("Double active function")
	}

	srv.active = handler
}

func (srv *HAServer) RegisterPassiveHandler(handler func() error) {
	if srv.passive != nil {
		log.Fatalln("Double passice function")
	}

	srv.passive = handler
}

func (srv *HAServer) NewActive(handler func() error) {
	if srv.active != nil {
		panic("Double Active")
	}
	srv.active = handler
}

func (srv *HAServer) NewPassive(handler func() error) {
	if srv.passive != nil {
		panic("Double Passive")
	}
	srv.passive = handler
}

func (srv *HAServer) SetVerbose(verbose bool) {
	srv.Reactor.SetVerbose(verbose)
}

func (srv *HAServer) StartReactor() error {
	if srv.voter == nil {
		log.Fatalln("Missing voter function")
	}

	srv.updatePeerExpiry()

	return srv.Reactor.Run(Heartbeat / 5)
}
