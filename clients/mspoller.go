package clients

import (
	"fmt"
	"github.com/pebbe/zmq4"
	"log"
)

func MsPoller() {

	//read from task ventilator
	receiver, err := zmq4.NewSocket(zmq4.PULL)
	if err != nil {
		log.Fatalln(err)
	}
	defer receiver.Close()

	err = receiver.Connect("tcp://localhost:5557")
	if err != nil {
		log.Fatalln(err)
	}

	//reads from weather update
	subscriber, err := zmq4.NewSocket(zmq4.SUB)
	if err != nil {
		log.Fatalln(err)
	}
	defer subscriber.Close()

	err = subscriber.Connect("tcp://localhost:5556")
	if err != nil {
		log.Fatalln(err)
	}

	err = subscriber.SetSubscribe("95134")
	if err != nil {
		log.Fatalln(err)
	}

	poller := zmq4.NewPoller()
	poller.Add(receiver, zmq4.POLLIN)
	poller.Add(subscriber, zmq4.POLLIN)

	for {
		sockets, err := poller.Poll(-1)
		if err != nil {
			log.Println(err)
		}

		for _, socket := range sockets {
			switch s := socket.Socket; s {
			case receiver:
				task, _ := s.Recv(0)
				if err != nil {
					log.Println(err)
				}

				fmt.Printf("Got task: %s\n", task)

			case subscriber:
				update, _ := s.Recv(0)
				if err != nil {
					log.Println(err)
				}

				fmt.Println("Got weather update:", update)
				return
			}
		}

	}

}
