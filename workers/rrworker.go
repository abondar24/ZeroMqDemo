package workers

import (
	"fmt"
	"github.com/pebbe/zmq4"
	"log"
	"time"
)

func RequestReplyRworker() {
	responder, err := zmq4.NewSocket(zmq4.REP)
	if err != nil {
		log.Fatal(err)
	}
	defer responder.Close()

	err = responder.Connect("tcp://localhost:5560")
	if err != nil {
		log.Fatalln(err)
	}

	for {
		request, err := responder.Recv(0)
		if err != nil {
			log.Println(err)
		}
		fmt.Printf("Received request: [%s]\n", request)

		time.Sleep(time.Second)
		_, err = responder.Send("World", 0)
		if err != nil {
			log.Println(err)
		}

	}
}
