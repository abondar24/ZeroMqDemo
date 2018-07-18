package clients

import (
	"fmt"
	"github.com/pebbe/zmq4"
	"log"
	"time"
)

func SyncSub() {
	subscriber, err := zmq4.NewSocket(zmq4.SUB)
	if err != nil {
		log.Fatalln(err)
	}

	defer subscriber.Close()

	err = subscriber.Connect("tcp://localhost:5561")
	if err != nil {
		log.Fatalln(err)
	}

	err = subscriber.SetSubscribe("")
	if err != nil {
		log.Fatalln(err)
	}

	time.Sleep(time.Second)

	//synchronize with publisher
	syncclient, err := zmq4.NewSocket(zmq4.REQ)
	if err != nil {
		log.Fatalln(err)
	}

	defer syncclient.Close()
	err = syncclient.Connect("tcp://localhost:5562")
	if err != nil {
		log.Fatalln(err)
	}

	_, err = syncclient.Send("", 0)
	if err != nil {
		log.Fatalln(err)
	}

	_, err = syncclient.Recv(0)
	if err != nil {
		log.Fatalln(err)
	}

	updateNum := 0
	for {
		msg, err := subscriber.Recv(0)
		if err != nil {
			log.Fatalln(err)
		}

		if msg == "END" {
			break
		}

		updateNum++
	}

	fmt.Printf("Received %d updates\n", updateNum)
}
