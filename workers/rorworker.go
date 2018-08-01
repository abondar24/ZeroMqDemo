package workers

import (
	"fmt"
	"github.com/pebbe/zmq4"
	"log"
	"math/rand"
	"time"
)

const (
	HeartbeatLiveness = 3
	HeartbeatInterval = 1000 * time.Millisecond
	IntervalInit      = 1000 * time.Millisecond
	IntervalMax       = 32000 * time.Millisecond
	RRReady           = "\001"
	RRHeartbeat       = "\002"
)

func configureWorkerSocket() (*zmq4.Socket, *zmq4.Poller) {
	worker, err := zmq4.NewSocket(zmq4.DEALER)
	if err != nil {
		log.Fatalln(err)
	}

	worker.Connect("tcp://localhost:5556")

	fmt.Println("Worker ready")
	worker.Send(RRReady, 0)

	poller := zmq4.NewPoller()
	poller.Add(worker, zmq4.POLLIN)

	return worker, poller

}

func RobustReliableWorker() {
	worker, poller := configureWorkerSocket()

	liveness := HeartbeatLiveness
	interval := IntervalInit

	heartbeatAt := time.Tick(HeartbeatInterval)

	rand.Seed(time.Now().UnixNano())
	for cycles := 0; true; {
		sockets, err := poller.PollAll(HeartbeatInterval)
		if err != nil {
			log.Fatalln(err)
		}

		if len(sockets) == 1 {
			//  - 3-part envelope + content -> request
			//  - 1-part HEARTBEAT -> heartbeat
			msg, err := worker.RecvMessage(0)
			if err != nil {
				log.Fatalln(err)
			}

			if len(msg) == 3 {
				//get nsg and simulate problems
				cycles++
				if cycles > 3 && rand.Intn(5) == 0 {
					fmt.Println("Crash")
					break
				} else if cycles > 3 && rand.Intn(5) == 0 {
					fmt.Println("CPU overlad")
					time.Sleep(3 * time.Second)
				}
				fmt.Println("Normal reply")
				worker.SendMessage(msg)
				liveness = HeartbeatLiveness
				time.Sleep(time.Second)
			} else if len(msg) == 1 {
				//heartbeat
				if msg[0] == RRHeartbeat {
					liveness = HeartbeatLiveness
				} else {
					fmt.Printf("Invalid message: %q\n", msg)
				}
			} else {
				fmt.Printf("Invalid message: %q\n", msg)
			}
			interval = IntervalInit
		} else {
			liveness--
			if liveness == 0 {
				fmt.Println("Heartbeat failure, queue is not available")
				fmt.Println("Reconnecting in", interval)
				time.Sleep(interval)

				if interval < IntervalMax {
					interval = 2 * interval
				}
				worker, poller = configureWorkerSocket()
				liveness = HeartbeatLiveness
			}
		}

		select {
		case <-heartbeatAt:
			fmt.Println("Send worker heartbeat")
			worker.Send(RRHeartbeat, 0)
		default:

		}
	}
}
