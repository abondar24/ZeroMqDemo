package clients

import (
	"log"
	"github.com/pebbe/zmq4"
	"fmt"
	"time"
)

func MsReader(){


	//read from task ventilator
	receiver,err := zmq4.NewSocket(zmq4.PULL)
	if err!=nil{
		log.Fatal(err)
	}
	defer  receiver.Close()
	receiver.Connect("tcp://localhost:5557")

	//reads from weather update
	subscriber,err:=zmq4.NewSocket(zmq4.SUB)
	if err!=nil{
		log.Fatal(err)
	}
    defer subscriber.Close()
	subscriber.Connect("tcp://localhost:5556")
	subscriber.SetSubscribe("95134")


	//vent traffic has bigger priority
	for {

		for {
			task,err := receiver.Recv(zmq4.DONTWAIT)
			if err!=nil{
				break
			}
			fmt.Println("Got task:",task)
		}

		for {
			update,err := subscriber.Recv(zmq4.DONTWAIT)
			if err!=nil{
				break
			}
			fmt.Println("Got weather update:",update)
		}

		time.Sleep(time.Microsecond)
	}
}