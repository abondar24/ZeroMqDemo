package queues

import (
	"github.com/pebbe/zmq4"
	"log"
)

const WorkerReady = "\001"

func ReliableQueue() {

	frontend, err := zmq4.NewSocket(zmq4.ROUTER)
	if err != nil {
		log.Fatalln(err)
	}
	defer frontend.Close()
	frontend.Bind("tcp://*:5555")

	backend, err := zmq4.NewSocket(zmq4.ROUTER)
	if err != nil {
		log.Fatalln(err)
	}
	defer backend.Close()
	backend.Bind("tcp://*:5556")

	workerQueue := make([]string, 0)

	poller1 := zmq4.NewPoller()
	poller1.Add(backend, zmq4.POLLIN)

	poller2 := zmq4.NewPoller()
	poller2.Add(backend, zmq4.POLLIN)
	poller2.Add(frontend, zmq4.POLLIN)

	for {
		var sockets []zmq4.Polled

		if len(workerQueue) > 0 {
			sockets, err = poller2.Poll(-1)
			if err != nil {
				log.Fatalln(err)
			}
		} else {
			sockets, err = poller1.Poll(-1)
			if err != nil {
				log.Fatalln(err)
			}
		}

		for _, socket := range sockets {
			switch s := socket.Socket; s {
			case backend:
				msg, err := s.RecvMessage(0)
				if err != nil {
					log.Fatalln(err)
				}

				id, msg := unwrap(msg)
				workerQueue = append(workerQueue, id)

				if msg[0] != WorkerReady {
					frontend.SendMessage(msg)
				}

			case frontend:
				msg, err := s.RecvMessage(0)
				if err != nil {
					log.Fatalln(err)
				}

				backend.SendMessage(workerQueue[0], "", msg)
			}
		}

	}
}

func unwrap(msg []string) (head string, tail []string) {
	head = msg[0]
	if len(msg) > 1 && msg[1] == "" {
		tail = msg[2:]
	} else {
		tail = msg[1:]
	}

	return
}
