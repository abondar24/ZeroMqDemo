package clients

import (
	"fmt"
	"github.com/pebbe/zmq4"
	"log"
)

func EnvSub() {
	subscriber, err := zmq4.NewSocket(zmq4.SUB)
	if err != nil {
		log.Fatalln(err)
	}

	defer subscriber.Close()

	err = subscriber.Connect("tcp://localhost:5563")
	if err != nil {
		log.Fatalln(err)
	}

	err = subscriber.SetSubscribe("B")
	if err != nil {
		log.Fatalln(err)
	}

	for {
		address, err := subscriber.Recv(0)
		if err != nil {
			log.Fatalln(err)
		}

		contents, err := subscriber.Recv(0)
		if err != nil {
			log.Fatalln(err)
		}

		fmt.Printf("[%s] %s\n", address, contents)

	}
}
