package main

import (
	"fmt"
	"github.com/pebbe/zmq4"
	"log"
	"regexp"
)

func Identity() {
	sink, err := zmq4.NewSocket(zmq4.ROUTER)
	if err != nil {
		log.Fatalln(err)
	}

	defer sink.Close()

	sink.Bind("inproc://test")

	anonym, err := zmq4.NewSocket(zmq4.REQ)
	if err != nil {
		log.Fatalln(err)
	}

	defer anonym.Close()

	anonym.Connect("inproc://test")
	anonym.Send("ROUTER user a generated UUID", 0)
	dump(sink)

	identified, err := zmq4.NewSocket(zmq4.REQ)
	if err != nil {
		log.Fatalln(err)
	}

	identified.SetIdentity("PEER2")
	identified.Connect("inproc://test")
	identified.Send("ROUTER socket uses REQ identity", 0)
	dump(sink)

}

func dump(soc *zmq4.Socket) {
	fmt.Println("---------------------------")
	allChar := regexp.MustCompile("^[^[:cntrl:]]*$")
	for {
		message, err := soc.Recv(0)
		if err != nil {
			log.Println(err)
		}

		fmt.Printf("[%03d]", len(message))
		if allChar.MatchString(message) {
			fmt.Print(message)
		} else {
			for i := 0; i < len(message); i++ {
				fmt.Printf("%02X", message[i])
			}
		}
		fmt.Println()

		more, err := soc.GetRcvmore()
		if !more {
			break
		}
	}
}
