package servers

import (
	"fmt"
	"github.com/pebbe/zmq4"
	"log"
	"time"
)

func LastValueCache() {
	frontend, err := zmq4.NewSocket(zmq4.SUB)
	if err != nil {
		log.Fatalln(err)
	}
	frontend.Bind("tcp://*:5557")
	frontend.SetSubscribe("")

	backend, err := zmq4.NewSocket(zmq4.XPUB)
	if err != nil {
		log.Fatalln(err)
	}
	backend.Bind("tcp://*:5558")

	cache := make(map[string]string)

	poller := zmq4.NewPoller()
	poller.Add(frontend, zmq4.POLLIN)
	poller.Add(backend, zmq4.POLLIN)

	for {
		polled, err := poller.Poll(1000 * time.Millisecond)
		if err != nil {
			log.Fatalln(err)
		}

		for _, item := range polled {
			switch socket := item.Socket; socket {
			case frontend:
				msg, err := frontend.RecvMessage(0)
				if err != nil {
					log.Fatalln(err)
				}
				cache[msg[0]] = msg[1]
				backend.SendMessage(msg)

			case backend:
				msg, err := backend.RecvMessage(0)
				if err != nil {
					log.Fatalln(err)
				}

				//  Event is one byte 0=unsub or 1=sub, followed by topic
				frame := msg[0]
				if frame[0] == 1 {
					topic := frame[1:]
					fmt.Println("Sending cached topic", topic)
					prev, ok := cache[topic]
					if ok {
						backend.SendMessage(topic, prev)
					}
				}
			}
		}
	}

}
