package clients

import (
	"fmt"
	"github.com/pebbe/zmq4"
	"log"
)

func HwClient() {

	socket, err := zmq4.NewSocket(zmq4.REQ)
	if err != nil {
		log.Fatalln(err)
	}

	defer socket.Close()

	fmt.Println("Connecting to hwserver")
	err = socket.Connect("tcp://localhost:5555")
	if err != nil {
		log.Fatalln(err)
	}

	for i := 0; i < 10; i++ {
		msg := fmt.Sprintf("Message %d", i)
		_, err = socket.Send(msg, 0)
		if err != nil {
			log.Println(err)
		}

		fmt.Println("Sending ", msg)

		reply, err := socket.Recv(0)
		if err != nil {
			log.Println(err)
		}

		fmt.Println("Received ", string(reply))
	}
}
