package servers

import (
	"fmt"
	"github.com/pebbe/zmq4"
	"log"
)

const (
	SUBSCRIBERS = 2
)

func SyncPub() {

	context, err := zmq4.NewContext()
	if err != nil {
		log.Fatalln(err)
	}

	defer context.Term()

	publisher, err := context.NewSocket(zmq4.PUB)
	if err != nil {
		log.Fatalln(err)
	}

	defer publisher.Close()
	err = publisher.SetSndhwm(1100000)
	if err != nil {
		log.Fatalln(err)
	}

	err = publisher.Bind("tcp://*:5561")
	if err != nil {
		log.Fatalln(err)
	}

	syncservice, err := context.NewSocket(zmq4.REP)
	if err != nil {
		log.Fatalln(err)
	}

	defer syncservice.Close()
	err = syncservice.Bind("tcp://*:5562")

	fmt.Println("Waiting for Subscribers")
	for subscribers := 0; subscribers < SUBSCRIBERS; subscribers++ {
		//  - wait for synchronization request
		_, err = syncservice.Recv(0)
		if err != nil {
			log.Fatalln(err)
		}

		//  - send synchronization reply
		_, err = syncservice.Send("", 0)
		if err != nil {
			log.Fatalln(err)
		}
	}

	fmt.Println("Broadcasting messages")
	for updateNum := 0; updateNum < 1000000; updateNum++ {
		_, err = publisher.Send("Rhubarb", 0)
		if err != nil {
			log.Fatalln(err)
		}
	}

	_, err = publisher.Send("END", 0)
	if err != nil {
		log.Fatalln(err)
	}
}
