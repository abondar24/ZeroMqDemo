package main

import (
	"fmt"
	"github.com/pebbe/zmq4"
	"log"
	"math/rand"
	"time"
)

func subscriber() {
	subscr, err := zmq4.NewSocket(zmq4.SUB)
	if err != nil {
		log.Fatalln(err)
	}

	subscr.Connect("tcp://localhost:6001")
	subscr.SetSubscribe("A")
	subscr.SetSubscribe("B")
	defer subscr.Close()

	for count := 0; count < 5; count++ {
		_, err = subscr.RecvMessage(0)
		if err != nil {
			log.Fatalln(err)
		}
	}
}

func publisher() {
	publ, err := zmq4.NewSocket(zmq4.PUB)
	if err != nil {
		log.Fatalln(err)
	}

	publ.Bind("tcp://*:6000")

	for {
		s := fmt.Sprintf("%c-%05d", rand.Intn(10)+'A', rand.Intn(100000))
		_, err := publ.SendMessage(s)
		if err != nil {
			log.Fatalln(err)
		}

		time.Sleep(100 * time.Millisecond)
	}
}

func listener() {
	pipe, err := zmq4.NewSocket(zmq4.PAIR)
	if err != nil {
		log.Fatalln(err)
	}

	pipe.Bind("inproc://pipe")

	for {
		msg, err := pipe.RecvMessage(0)
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Printf("%q\n", msg)
	}
}

func PubSubTracing() {

	go publisher()
	go subscriber()
	go listener()

	time.Sleep(100 * time.Millisecond)

	subscr, err := zmq4.NewSocket(zmq4.XSUB)
	if err != nil {
		log.Fatalln(err)
	}
	subscr.Connect("tcp://localhost:6000")

	publ, err := zmq4.NewSocket(zmq4.XPUB)
	if err != nil {
		log.Fatalln(err)
	}
	publ.Bind("tcp://*:6001")

	listn, err := zmq4.NewSocket(zmq4.PAIR)
	if err != nil {
		log.Fatalln(err)
	}
	listn.Connect("inproc://pipe")
	zmq4.Proxy(subscr, publ, listn)

	fmt.Println("interrupted")

}
