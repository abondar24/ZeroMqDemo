package workers

import (
	"fmt"
	"github.com/pebbe/zmq4"
	"log"
	"strconv"
	"time"
)

func TaskWorker() {

	//get msg from ventilator
	receiver, err := zmq4.NewSocket(zmq4.PULL)
	if err != nil {
		log.Fatalln(err)
	}
	defer receiver.Close()

	err = receiver.Connect("tcp://localhost:5557")
	if err != nil {
		log.Fatalln(err)
	}

	//send messages to task sink
	sender, err := zmq4.NewSocket(zmq4.PUSH)
	if err != nil {
		log.Fatal(err)
	}
	defer sender.Close()

	err = sender.Connect("tcp://localhost:5558")
	if err != nil {
		log.Fatalln(err)
	}

	//input control socket
	controller, err := zmq4.NewSocket(zmq4.SUB)
	if err != nil {
		log.Fatal(err)
	}
	defer controller.Close()

	err = controller.Connect("tcp://localhost:5559")
	if err != nil {
		log.Fatalln(err)
	}

	err = controller.SetSubscribe("")
	if err != nil {
		log.Fatalln(err)
	}

	poller := zmq4.NewPoller()
	poller.Add(receiver, zmq4.POLLIN)
	poller.Add(controller, zmq4.POLLIN)
	// process tasks
	for {

		sockets, err := poller.Poll(-1)
		if err != nil {
			log.Println(err)
		}

		for _, socket := range sockets {
			switch s := socket.Socket; s {
			case receiver:
				msg, _ := s.Recv(0)
				if err != nil {
					log.Println(err)
				}

				fmt.Printf("%s.", msg)

				// Do the work
				msec, err := strconv.ParseInt(msg, 10, 64)
				if err != nil {
					log.Println(err)
				}

				time.Sleep(time.Duration(msec) * 1e6)

				_, err = sender.Send(msg, 0)
				if err != nil {
					log.Println(err)
				}

			case controller:
				fmt.Println("stopping")
				return
			}
		}
	}
}
