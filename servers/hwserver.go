package servers

import (
	"fmt"
	"github.com/pebbe/zmq4"
	"log"
	"time"
)

func HwServer() {

	socket, err := zmq4.NewSocket(zmq4.REP)
	if err != nil {
		log.Fatalln(err)
	}

	defer socket.Close()

	err = socket.Bind("tcp://*:5555")
	if err != nil {
		log.Fatalln(err)
	}

	for {
		msg, err := socket.Recv(0)
		if err != nil {
			log.Println(msg)
		}

		fmt.Println("Received a message ", string(msg))

		time.Sleep(time.Second)

		reply := fmt.Sprintf("Hi")
		_, err = socket.Send(reply, 0)
		if err != nil {
			log.Println(msg)
		}
	}

}
