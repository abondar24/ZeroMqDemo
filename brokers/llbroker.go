package brokers

import (
	"fmt"
	"github.com/pebbe/zmq4"
	"log"
	"strings"
	"time"
)

const (
	NumberOfClients = 10
	NumberOfWorkers = 3
	WorkerReady     = "\001"
)

//client
func clientTask() {
	client, err := zmq4.NewSocket(zmq4.REQ)
	if err != nil {
		log.Fatalln(err)
	}

	defer client.Close()
	client.Connect("ipc://frontend.ipc")

	for {
		client.SendMessage("Hello")
		reply, err := client.RecvMessage(0)
		if err != nil {
			log.Println(err)
		}

		fmt.Println("Client:", strings.Join(reply, "\n\t"))
		time.Sleep(time.Second)
	}

}

//worker uses req for load-balance
func workerTask() {
	worker, err := zmq4.NewSocket(zmq4.REQ)
	if err != nil {
		log.Fatalln(err)
	}

	defer worker.Close()
	worker.Connect("ipc://backend.ipc")
	worker.SendMessage(WorkerReady)

	for {

		msg, err := worker.RecvMessage(0)
		if err != nil {
			log.Fatalln(err)
		}

		msg[len(msg)-1] = "OK"
		worker.SendMessage(msg)
	}

}

func LoadBalacningBroker() {

	frontend, err := zmq4.NewSocket(zmq4.ROUTER)
	if err != nil {
		log.Fatalln(err)
	}
	defer frontend.Close()
	frontend.Bind("ipc://frontend.ipc")

	backend, err := zmq4.NewSocket(zmq4.ROUTER)
	if err != nil {
		log.Fatalln(err)
	}
	defer backend.Close()
	backend.Bind("ipc://backend.ipc")

	clientsNum := 0
	for ; clientsNum < NumberOfClients; clientsNum++ {
		go clientTask()
	}

	for workerNum := 0; workerNum < NumberOfWorkers; workerNum++ {
		go workerTask()
	}

	workerQueue := make([]string, 0, 10)

	pollerBackEnd := zmq4.NewPoller()
	pollerBackEnd.Add(backend, zmq4.POLLIN)

	pollerFrontEnd := zmq4.NewPoller()
	pollerFrontEnd.Add(backend, zmq4.POLLIN)
	pollerFrontEnd.Add(frontend, zmq4.POLLIN)

	for {
		var sockets []zmq4.Polled
		// poll backend always, frontend - when workers > 0

		if len(workerQueue) > 0 {
			sockets, err = pollerFrontEnd.Poll(-1)
			if err != nil {
				log.Fatalln(err)
			}
		} else {
			sockets, err = pollerBackEnd.Poll(-1)
			if err != nil {
				log.Fatalln(err)
			}

		}

		for _, socket := range sockets {
			switch socket.Socket {
			case backend:
				msg, err := backend.RecvMessage(0)
				if err != nil {
					log.Fatalln(err)
				}

				id, msg := unwrap(msg)
				workerQueue = append(workerQueue, id)

				fmt.Println(msg)
				if msg[0] != WorkerReady {
					frontend.SendMessage(msg)
				}

			case frontend:
				msg, err := frontend.RecvMessage(0)
				if err != nil {
					log.Println(err)
				} else {
					fmt.Println(msg)
					backend.SendMessage(workerQueue[0], "", msg)

					// Deque and dropm the next worker
					workerQueue = workerQueue[1:]
				}

			}

		}
	}

	time.Sleep(100 * time.Millisecond)

}

func unwrap(msg []string) (head string, tail []string) {
	head = msg[0]
	if len(msg) > 1 && msg[1] == "" {
		tail = msg[2:]
	} else {
		tail = msg[1:]
	}

	return
}
