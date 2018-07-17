package workers

import (
	"github.com/pebbe/zmq4"
	"log"
	"fmt"
	"time"
)

func RRworker(){
	responder,err := zmq4.NewSocket(zmq4.REP)
	if err!=nil{
		log.Fatal(err)
	}
	defer responder.Close()
	responder.Connect("tcp://localhost:5560")

	for {
		request,err := responder.Recv(0)
		if err!=nil{
			log.Println(err)
		}
		fmt.Printf("Received request: [%s]\n",request)

		time.Sleep(time.Second)
		responder.Send("World",0)

	}
}
