package servers

import (
	"fmt"
	"github.com/pebbe/zmq4"
	"log"
	"math/rand"
	"time"
)

func TaskVentilator() {

	sender, err := zmq4.NewSocket(zmq4.PUSH)
	if err != nil {
		log.Fatalln(err)
	}
	defer sender.Close()

	err = sender.Bind("tcp://*:5557")
	if err != nil {
		log.Fatalln(err)
	}

	//sender for sending start of batch message
	sink, err := zmq4.NewSocket(zmq4.PUSH)
	if err != nil {
		log.Fatalln(err)
	}
	defer sink.Close()

	err = sink.Connect("tcp://localhost:5558")
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Print("Print Enter when workers are ready:")

	var line string
	fmt.Scanln(&line)

	fmt.Println("Sending tasks to workers")

	_, err = sink.Send("0", 0)
	if err != nil {
		log.Fatalln(err)
	}

	rand.Seed(time.Now().UnixNano())

	totalMsec := 0

	for i := 0; i < 100; i++ {
		workload := rand.Intn(100)
		totalMsec += workload
		msg := fmt.Sprintf("%d", workload)
		_, err = sender.Send(msg, 0)
		if err != nil {
			log.Println()
		}
	}

	fmt.Printf("Total expected cost: %d msec\n", totalMsec)

	//time to deliver: 1sec
	time.Sleep(5e9)
}
