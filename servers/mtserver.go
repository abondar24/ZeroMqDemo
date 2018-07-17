package servers

import (
	"fmt"
	"github.com/pebbe/zmq4"
	"log"
	"time"
)

func MTserver() {
	clients, err := zmq4.NewSocket(zmq4.ROUTER)
	if err != nil {
		log.Fatalln(err)
	}

	defer clients.Close()

	err = clients.Bind("tcp://*:5555")
	if err != nil {
		log.Fatalln(err)
	}

	workers, err := zmq4.NewSocket(zmq4.DEALER)
	if err != nil {
		log.Fatalln(err)
	}

	defer workers.Close()

	err = workers.Bind("inproc://workers")
	if err != nil {
		log.Fatalln(err)
	}

	//launch pool of goroutines
	for threadNum := 0; threadNum < 5; threadNum++ {
		go workRoutine()
	}

	err = zmq4.Proxy(clients, workers, nil)
	if err != nil {
		log.Fatalln(err)
	}

}

func workRoutine() {
	//socket to to talk to dispatcher
	receiver, err := zmq4.NewSocket(zmq4.REP)
	if err != nil {
		log.Fatalln(err)
	}

	defer receiver.Close()

	err = receiver.Connect("inproc://workers")
	if err != nil {
		log.Fatalln(err)
	}

	for {
		msg, err := receiver.Recv(0)
		if err != nil {
			break
		}
		fmt.Println("Received request: [" + msg + "]")

		time.Sleep(time.Second)

		_, err = receiver.Send("World", 0)
		if err != nil {
			break
		}
	}

}
