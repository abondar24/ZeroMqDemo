package servers

import (
	"github.com/pebbe/zmq4"
	"log"
	"fmt"
	"time"
)

func HwServer(){
    context,err :=zmq4.NewContext()

    if err!=nil{
		log.Fatal(err)
	}

	socket, err := context.NewSocket(zmq4.REP)
	if err!=nil{
		log.Fatal(err)
	}


	defer  context.Term()
	defer  socket.Close()

	socket.Bind("tcp://*:5555")

	for {
		msg,err := socket.Recv(0)
		if err!=nil{
			log.Println(msg)
		}

		fmt.Println("Received a message ", string(msg))

		time.Sleep(time.Second)

		reply := fmt.Sprintf("Hi")
		socket.Send(reply,0)
	}

}