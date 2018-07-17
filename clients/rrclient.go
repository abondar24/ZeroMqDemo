package clients

import (
	"fmt"
	"github.com/pebbe/zmq4"
	"log"
)

func RRclient() {
	requester, err := zmq4.NewSocket(zmq4.REQ)
	if err != nil {
		log.Fatalln(err)
	}
	defer requester.Close()

	err = requester.Connect("tcp://localhost:5559")
	if err != nil {
		log.Fatalln(err)
	}

	for request := 0; request < 10; request++ {
		_, err = requester.Send("Hello", 0)
		if err != nil {
			log.Println(err)
		}

		reply, err := requester.Recv(0)
		if err != nil {
			log.Println(err)
		}
		fmt.Printf("Received reply %d [%s]\n", request, reply)
	}

}
