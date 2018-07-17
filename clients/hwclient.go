package clients

import (
	"log"
	"fmt"
	"github.com/pebbe/zmq4"
)

func HwClient(){

	socket, err := zmq4.NewSocket(zmq4.REQ)
	if err!=nil{
		log.Fatal(err)
	}

	defer  socket.Close()

	fmt.Println("Connecting to hwserver")
	socket.Connect("tcp://localhost:5555")

	for i:=0; i<10; i++{
		msg := fmt.Sprintf("Message %d",i)
		socket.Send(msg,0)
		fmt.Println("Sending ", msg)

		reply,err :=socket.Recv(0)
		if err!=nil{
			log.Println(err)
		}

		fmt.Println("Received ",string(reply))
	}
}