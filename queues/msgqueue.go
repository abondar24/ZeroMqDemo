package queues

import (
	"github.com/pebbe/zmq4"
	"log"
)

func MsgQueue() {
	frontend, err := zmq4.NewSocket(zmq4.ROUTER)
	if err != nil {
		log.Fatalln(err)
	}
	defer frontend.Close()

	err = frontend.Bind("tcp://*:5559")
	if err != nil {
		log.Fatalln(err)
	}

	backend, err := zmq4.NewSocket(zmq4.DEALER)
	if err != nil {
		log.Fatalln(err)
	}

	defer backend.Close()

	err = backend.Bind("tcp://*:5560")
	if err != nil {
		log.Fatalln(err)
	}

	err = zmq4.Proxy(frontend, backend, nil)
	if err != nil {
		log.Fatalln(err)
	}

}
