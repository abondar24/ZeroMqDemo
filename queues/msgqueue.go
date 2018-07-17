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

	if frontend.Bind("tcp://*:5559") != nil {
		log.Fatalln(err)
	}

	backend, err := zmq4.NewSocket(zmq4.DEALER)
	if err != nil {
		log.Fatalln(err)
	}

	defer backend.Close()
	if backend.Bind("tcp://*:5560") != nil {
		log.Fatalln(err)
	}

	if zmq4.Proxy(frontend,backend,nil)!=nil{
		log.Fatalln(err)
	}

}
