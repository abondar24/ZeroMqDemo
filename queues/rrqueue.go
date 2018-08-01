package queues

import (
	"fmt"
	"github.com/pebbe/zmq4"
	"log"
	"time"
)

const (
	HeartbeatLiveness = 3
	HeartBeatInterval = 1000 * time.Millisecond

	RRReady     = "\001"
	RRHeartbeat = "\002"
)

type Worker struct {
	id     string
	idStr  string
	expire time.Time
}

func NewWorker(id string) Worker {
	return Worker{
		id:     id,
		idStr:  id,
		expire: time.Now().Add(HeartBeatInterval * HeartbeatLiveness),
	}
}

func PrepareWorker(self Worker, workers []Worker) []Worker {
	for i, worker := range workers {
		if self.idStr == worker.idStr {
			if i == 0 {
				workers = workers[1:]
			} else if i == len(workers)-1 {
				workers = workers[:i]
			} else {
				workers = append(workers[:i], workers[i+1:]...)
			}
			break
		}
	}
	return append(workers, self)
}

func PurgeWorkers(workers []Worker) []Worker {
	now := time.Now()
	for i, worker := range workers {
		if now.Before(worker.expire) {
			return workers[i:]
		}
	}

	return workers[0:0]
}

func RobustReliableQueue() {
	frontend, err := zmq4.NewSocket(zmq4.ROUTER)
	if err != nil {
		log.Fatalln(err)
	}
	defer frontend.Close()
	frontend.Bind("tcp://*:5555")

	backend, err := zmq4.NewSocket(zmq4.ROUTER)
	if err != nil {
		log.Fatalln(err)
	}
	defer backend.Close()
	backend.Bind("tcp://*:5556")

	workerQueue := make([]Worker, 0)

	heartbeatAt := time.Tick(HeartBeatInterval)

	poller1 := zmq4.NewPoller()
	poller1.Add(backend, zmq4.POLLIN)

	poller2 := zmq4.NewPoller()
	poller2.Add(backend, zmq4.POLLIN)
	poller2.Add(frontend, zmq4.POLLIN)

	for {
		var sockets []zmq4.Polled

		if len(workerQueue) > 0 {
			sockets, err = poller2.Poll(HeartBeatInterval)
			if err != nil {
				log.Fatalln(err)
			}
		} else {
			sockets, err = poller1.Poll(HeartBeatInterval)
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

				id, msg := unwrapMsg(msg)
				workerQueue = PrepareWorker(NewWorker(id), workerQueue)

				if len(msg) == 1 {
					if msg[0] != RRReady && msg[0] != RRHeartbeat {
						fmt.Println("Invalid message from worker", msg)
					}
				} else {
					frontend.SendMessage(msg)
				}

			case frontend:
				msg, err := frontend.RecvMessage(0)
				if err != nil {
					log.Fatalln(err)
				}

				backend.SendMessage(workerQueue[0].id, msg)
				workerQueue = workerQueue[1:]
			}
		}

		//handle heartbeat after every socket activity
		select {
		case <-heartbeatAt:
			for _, worker := range workerQueue {
				backend.SendMessage(worker.id, RRHeartbeat)
			}
		default:

		}
		workerQueue = PurgeWorkers(workerQueue)
	}
}

func unwrapMsg(msg []string) (head string, tail []string) {
	head = msg[0]
	if len(msg) > 1 && msg[1] == "" {
		tail = msg[2:]
	} else {
		tail = msg[1:]
	}

	return
}
