package main

import (
	"fmt"
	"github.com/pebbe/zmq4"
	"log"
	"math/rand"
	"time"
)

const NumberOfWorkers = 10

func workerTask() {
	//use dealer for async
	worker, err := zmq4.NewSocket(zmq4.REQ)
	if err != nil {
		log.Fatalln(err)
	}
	defer worker.Close()
	setIdentity(worker)
	worker.Connect("tcp://localhost:5671")

	total := 0
	for {
		worker.Send("Hiii", 0)

		workload, err := worker.Recv(0)
		if err != nil {
			log.Println(err)
		} else if workload == "Fired!" {
			fmt.Printf("Completed: %d tasks\n", total)
			break
		}

		total++

		time.Sleep(time.Duration(rand.Intn(500)+1) * time.Microsecond)
	}

}

func setIdentity(soc *zmq4.Socket) {
	id := fmt.Sprintf("%04X-%04X", rand.Intn(0x100000), rand.Intn(0x100000))
	soc.SetIdentity(id)
}

func ReqRouter() {
	broker, err := zmq4.NewSocket(zmq4.ROUTER)
	if err != nil {
		log.Fatalln(err)
	}

	defer broker.Close()
	broker.Bind("tcp://*:5671")

	rand.Seed(time.Now().UnixNano())
	for workerNum := 0; workerNum < NumberOfWorkers; workerNum++ {
		go workerTask()
	}

	startTime := time.Now()
	workersFired := 0

	for {
		id, err := broker.Recv(0)
		if err != nil {
			log.Println(err)
		}

		broker.Send(id, zmq4.SNDMORE)
		broker.Recv(0) //envelope delimitter
		broker.Recv(0) //response from worker
		broker.Send("", zmq4.SNDMORE)

		if time.Since(startTime) < 5*time.Second {
			broker.Send("Keep Pushing", 0)
		} else {
			broker.Send("Fired!", 0)
			workersFired++
			if workersFired == NumberOfWorkers {
				break
			}
		}
	}

	time.Sleep(time.Second)
}
