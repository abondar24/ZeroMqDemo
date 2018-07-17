package brokers

import (
	"github.com/pebbe/zmq4"
	"log"
)

func RRbroker() {
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

	poller := zmq4.NewPoller()
	poller.Add(frontend, zmq4.POLLIN)
	poller.Add(backend, zmq4.POLLIN)

	for {
		sockets, err := poller.Poll(-1)
		if err != nil {
			log.Println(err)
		}

		for _, socket := range sockets {
			switch s := socket.Socket; s {
			case frontend:
				for {
					msg, err := s.Recv(0)
					if err != nil {
						log.Println(err)
					}

					if more, _ := s.GetRcvmore(); more {
						_, err = backend.Send(msg, zmq4.SNDMORE)
						if err != nil {
							log.Println(err)
						}
					} else {
						_, err = backend.Send(msg, 0)
						if err != nil {
							log.Println(err)
						}
						break
					}
				}
			case backend:
				for {
					msg, err := s.Recv(0)
					if err != nil {
						log.Println(err)
					}

					if more, _ := s.GetRcvmore(); more {
						_, err = frontend.Send(msg, zmq4.SNDMORE)
						if err != nil {
							log.Println(err)
						}
					} else {
						_, err = frontend.Send(msg, 0)
						if err != nil {
							log.Println(err)
						}
						break
					}
				}
			}
		}

	}
}
