package clients

import (
	"fmt"
	"github.com/pebbe/zmq4"
	"log"
	"time"
)

func MsReader() {

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

	//vent traffic has bigger priority
	for {

		for {
			task, err := receiver.Recv(zmq4.DONTWAIT)
			if err != nil {
				break
			}
			fmt.Println("Got task:", task)
		}

		for {
			update, err := subscriber.Recv(zmq4.DONTWAIT)
			if err != nil {
				break
			}
			fmt.Println("Got weather update:", update)
		}

		time.Sleep(time.Microsecond)
	}
}
