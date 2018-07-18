package main

import (
	"fmt"
	"github.com/pebbe/zmq4"
	"log"
)

func step1() {
	xmitter, err := zmq4.NewSocket(zmq4.PAIR)
	if err != nil {
		log.Fatalln(err)
	}

	defer xmitter.Close()
	err = xmitter.Connect("inproc://step2")
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println("Step 1 ready,signalling step 2")
	_, err = xmitter.Send("READY", 0)
	if err != nil {
		log.Fatalln(err)
	}
}

func step2() {
	receiver, err := zmq4.NewSocket(zmq4.PAIR)
	if err != nil {
		log.Fatalln(err)
	}

	defer receiver.Close()
	err = receiver.Bind("inproc://step2")
	if err != nil {
		log.Fatalln(err)
	}

	go step1()

	_, err = receiver.Recv(0)
	if err != nil {
		log.Fatalln(err)
	}

	step3()

}

func step3() {
	xmitter, err := zmq4.NewSocket(zmq4.PAIR)
	if err != nil {
		log.Fatalln(err)
	}

	defer xmitter.Close()
	err = xmitter.Connect("inproc://step3")
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println("Step 2 ready, signaling step 3")
	_, err = xmitter.Send("READY", 0)
	if err != nil {
		log.Fatalln(err)
	}
}

func MTrelay() {
	receiver, err := zmq4.NewSocket(zmq4.PAIR)
	if err != nil {
		log.Fatalln(err)
	}

	defer receiver.Close()

	err = receiver.Bind("inproc://step3")
	if err != nil {
		log.Fatalln(err)
	}

	go step2()

	_, err = receiver.Recv(0)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println("Test Successful!")

}
