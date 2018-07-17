package clients

import (
	"fmt"
	"github.com/pebbe/zmq4"
	"log"
	"time"
)

func TaskSink() {

	//get msg from ventilator
	receiver, err := zmq4.NewSocket(zmq4.PULL)
	if err != nil {
		log.Fatalln(err)
	}
	defer receiver.Close()

	err = receiver.Bind("tcp://*:5558")
	if err != nil {
		log.Fatalln(err)
	}

	//input control socket
	controller, err := zmq4.NewSocket(zmq4.PUB)
	if err != nil {
		log.Fatalln(err)
	}
	defer controller.Close()

	err = controller.Bind("tcp://*:5559")
	if err != nil {
		log.Fatalln(err)
	}

	//wait batch to start
	msg, err := receiver.Recv(0)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println("Received start msg", msg)

	startTime := time.Now().UnixNano()

	for i := 0; i < 100; i++ {
		msg, err = receiver.Recv(0)
		fmt.Println(msg)
		if err != nil {
			log.Println(err)
		}

		if i%10 == 0 {
			fmt.Println(":")
		} else {
			fmt.Println(".")
		}
	}

	timeEnd := time.Now().UnixNano()
	fmt.Printf("Total elapsed time: %d msec\n", (timeEnd-startTime)/1e6)

	_, err = controller.Send("KILL", 0)
	if err != nil {
		log.Println(err)
	}

	time.Sleep(1 * time.Second)

}
