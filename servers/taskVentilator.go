package servers

import (
	"log"
	"fmt"
	"github.com/pebbe/zmq4"
	"math/rand"
	"time"
)

func TaskVentilator(){


	sender, err := zmq4.NewSocket(zmq4.PUSH)
	if err!=nil{
		log.Fatal(err)
	}
	defer  sender.Close()

	sender.Bind("tcp://*:5557")

	//sender for sending start of batch message
	sink,err := zmq4.NewSocket(zmq4.PUSH)
	if err!=nil{
		log.Fatal(err)
	}
	defer sink.Close()
	sink.Connect("tcp://localhost:5558")

	fmt.Print("Print Enter when workers are ready:")

	var line string
	fmt.Scanln(&line)

	fmt.Println("Sending tasks to workers")

	sink.Send("0",0)

	rand.Seed(time.Now().UnixNano())

	totalMsec := 0

	for i:=0;i<100;i++{
		workload := rand.Intn(100)
		totalMsec +=workload
		msg := fmt.Sprintf("%d", workload)
		sender.Send(msg,0)
	}

	fmt.Printf("Total expected cost: %d msec\n",totalMsec)

	//time to deliver: 1sec
	time.Sleep(5e9)
}
