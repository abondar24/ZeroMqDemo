package clients

import (
	"log"
	"github.com/pebbe/zmq4"
	"fmt"
)

func RRclient(){
	requester,err := zmq4.NewSocket(zmq4.REQ)
	if err!=nil{
		log.Fatal(err)
	}
	defer requester.Close()
	requester.Connect("tcp://localhost:5559")

	for request :=0; request<10;request++{
		requester.Send("Hello",0)
		reply,err := requester.Recv(0)
		if err!=nil{
			log.Println(err)
		}
		fmt.Printf("Received reply %d [%s]\n",request,reply)
	}

}