package mdapi

import (
	"errors"
	"github.com/pebbe/zmq4"
	"log"
	"runtime"
	"time"
)

type MdClient struct {
	broker  string
	client  *zmq4.Socket
	verbose bool
	timeout time.Duration
	poller  *zmq4.Poller
}

var (
	errPermanent = errors.New("permanent error, abandoning request")
)

func (mdclient *MdClient) ConnectToBroker() (err error) {
	if mdclient.client != nil {
		mdclient.client.Close()
		mdclient.client = nil
	}

	mdclient.client, err = zmq4.NewSocket(zmq4.DEALER)
	if err != nil {
		if mdclient.verbose {
			log.Println("Socket creating failed")
		}
		return
	}

	mdclient.poller = zmq4.NewPoller()
	mdclient.poller.Add(mdclient.client, zmq4.POLLIN)

	if mdclient.verbose {
		log.Printf("Connecting to broker at %s...", mdclient.broker)
	}

	err = mdclient.client.Connect(mdclient.broker)
	if err != nil && mdclient.verbose {
		log.Println("Broker connection failed ", mdclient.broker)
	}

	return
}

func NewMdClient(broker string, verbose bool) (mdclient *MdClient, err error) {
	mdclient = &MdClient{
		broker:  broker,
		verbose: verbose,
		timeout: time.Duration(2500 * time.Millisecond),
	}

	err = mdclient.ConnectToBroker()
	runtime.SetFinalizer(mdclient, (*MdClient).Close)
	return
}

func (mdclient *MdClient) Close() (err error) {
	if mdclient.client != nil {
		err = mdclient.client.Close()
		mdclient.client = nil
	}

	return
}

func (mdclient *MdClient) SetTimeout(timeout time.Duration) {
	mdclient.timeout = timeout
}

func (mdclient *MdClient) Send(service string, request ...string) (err error) {

	//  Prefix request with protocol frames
	//  Frame 0: empty (REQ emulation)
	//  Frame 1: "MDPCxy" (six bytes, MDP/Client x.y)
	//  Frame 2: Service name (printable string)

	req := make([]string, 3, len(request)+3)
	req = append(req, request...)
	req[2] = service
	req[1] = MdClientVer
	req[0] = ""
	if mdclient.verbose {
		log.Printf("Send request to '%s' service: %q\n", service, req)
	}

	_, err = mdclient.client.SendMessage(req)
	return
}

func (mdclient *MdClient) Recv() (msg []string, err error) {
	msg = []string{}

	polled, err := mdclient.poller.PollAll(mdclient.timeout)
	if err != nil {
		return
	}

	if len(polled) > 0 {
		msg, err = mdclient.client.RecvMessage(0)
		if err != nil {
			log.Println("Interrupt received")
			log.Println(err)
			return
		}

		if mdclient.verbose {
			log.Println("Received reply: %q\n", msg)
		}

		if len(msg) < 4 {
			log.Fatalln("len(msg) < 4")
		}

		if msg[0] != "" {
			log.Fatalln("msg[0] !=\"\"")
		}

		if msg[1] != MdClientVer {
			log.Fatalln("msg[1] != MdClientVer")
		}

		msg = msg[3:]
		return
	}

	if mdclient.verbose {
		log.Println(errPermanent)
	}
	return
}
