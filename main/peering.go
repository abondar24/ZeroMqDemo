package main

import (
	"fmt"
	"github.com/pebbe/zmq4"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

const (
	NumberClients = 10
	NumberWorkers = 5
	WorkerReady   = "**READY**"
)

var self string

//issues requests and then sleeps for a while.
func peeringClient(i int) {

	client, err := zmq4.NewSocket(zmq4.REQ)
	if err != nil {
		log.Fatalln(err)
	}
	defer client.Close()
	client.Connect("ipc://" + self + "-localfe.ipc")

	monitor, err := zmq4.NewSocket(zmq4.PUSH)
	if err != nil {
		log.Fatalln(err)
	}
	defer monitor.Close()
	monitor.Connect("ipc://" + self + "-monitor.ipc")

	poller := zmq4.NewPoller()
	poller.Add(client, zmq4.POLLIN)

	for {
		time.Sleep(time.Duration(rand.Intn(5000)) * time.Millisecond)

		for burst := rand.Intn(15); burst > 0; burst-- {
			taskId := fmt.Sprintf("%04X-%s-%d", rand.Intn(0x10000), self, i)

			//send request with random id
			client.Send(taskId, 0)

			//wait for reply
			sockets, err := poller.PollAll(10 * time.Second)
			if err != nil {
				log.Fatalln(err)
			}

			if len(sockets) == 1 {
				reply, err := client.Recv(0)
				if err != nil {
					log.Fatalln(err)
				}

				id := strings.Fields(reply)[0]

				if id != taskId {
					log.Fatalln("id!=taskId")
				}
				monitor.Send(reply, 0)
			} else {
				monitor.Send("E: CLIENT EXIT - lost task "+taskId, 0)
				return
			}

		}
	}
}

//req - for load-balancer.
func peeringWorker(i int) {
	worker, err := zmq4.NewSocket(zmq4.REQ)
	if err != nil {
		log.Fatalln(err)
	}

	defer worker.Close()
	worker.Connect("ipc://" + self + "-localbe.ipc")

	worker.SendMessage(WorkerReady)

	for {
		msg, err := worker.RecvMessage(0)
		if err != nil {
			log.Fatalln(err)
		}

		time.Sleep(time.Duration(rand.Intn(2)) * time.Second)
		n := len(msg) - 1
		worker.SendMessage(msg[:n], fmt.Sprintf("%s %s-%d", msg[n], self, i))
	}
}

func Peering(brokerName string, peers ...string) {

	self = brokerName
	fmt.Printf("Preparing broker at %s..\n", self)
	rand.Seed(time.Now().UnixNano())

	localfe, err := zmq4.NewSocket(zmq4.ROUTER)
	if err != nil {
		log.Fatalln(err)
	}
	defer localfe.Close()
	localfe.Bind("ipc://" + self + "-localfe.ipc")

	localbe, err := zmq4.NewSocket(zmq4.ROUTER)
	if err != nil {
		log.Fatalln(err)
	}
	defer localbe.Close()
	localbe.Bind("ipc://" + self + "-localbe.ipc")

	cloudfe, err := zmq4.NewSocket(zmq4.ROUTER)
	if err != nil {
		log.Fatalln(err)
	}
	defer cloudfe.Close()

	cloudfe.SetIdentity(self)
	cloudfe.Bind("ipc://" + self + "-cloud.ipc")

	cloudbe, err := zmq4.NewSocket(zmq4.ROUTER)
	if err != nil {
		log.Fatalln(err)
	}

	defer cloudbe.Close()
	cloudbe.SetIdentity(self)

	for _, peer := range peers {
		fmt.Printf("Connecting to cloud frontend at '%s'\n", peer)
		cloudbe.Connect("ipc://" + self + "-cloud.ipc")
	}

	statebe, err := zmq4.NewSocket(zmq4.PUB)
	if err != nil {
		log.Fatalln(err)
	}
	defer statebe.Close()
	statebe.Bind("ipc://" + self + "-state.ipc")

	statefe, err := zmq4.NewSocket(zmq4.SUB)
	if err != nil {
		log.Fatalln(err)
	}
	defer statefe.Close()
	statefe.SetSubscribe("")

	for _, peer := range peers {
		fmt.Printf("Connecting to state backend at '%s'\n", peer)
		statefe.Connect("ipc://" + self + "-state.ipc")
	}

	monitor, err := zmq4.NewSocket(zmq4.PULL)
	if err != nil {
		log.Fatalln(err)
	}
	defer monitor.Close()
	monitor.Bind("ipc://" + self + "-monitor.ipc")

	for workerNum := 0; workerNum < NumberWorkers; workerNum++ {
		go peeringWorker(workerNum)
	}

	for clientNum := 0; clientNum < NumberClients; clientNum++ {
		go peeringClient(clientNum)
	}

	localCapacity := 0
	cloudCapacity := 0

	workerQueue := make([]string, 0)

	primary := zmq4.NewPoller()
	primary.Add(localbe, zmq4.POLLIN)
	primary.Add(cloudbe, zmq4.POLLIN)
	primary.Add(statefe, zmq4.POLLIN)
	primary.Add(monitor, zmq4.POLLIN)

	secondary1 := zmq4.NewPoller()
	secondary1.Add(localfe, zmq4.POLLIN)

	secondary2 := zmq4.NewPoller()
	secondary2.Add(localfe, zmq4.POLLIN)
	secondary2.Add(cloudfe, zmq4.POLLIN)

	msg := make([]string, 0)

	for {
		timeout := time.Duration(time.Second)
		if localCapacity == 0 {
			timeout = -1
		}

		sockets, err := primary.PollAll(timeout)
		if err != nil {
			log.Fatalln(err)
		}

		previousCapacity := localCapacity

		//reply from localworker
		msg = msg[0:0]

		//0 is localbe
		if sockets[0].Events&zmq4.POLLIN != 0 {
			msg, err = localbe.RecvMessage(0)
			if err != nil {
				log.Fatalln(err)
			}

			id, msg := unwrap(msg)
			workerQueue = append(workerQueue, id)
			localCapacity++

			if msg[0] == WorkerReady {
				msg = msg[0:0]
			}

		} else if sockets[1].Events&zmq4.POLLIN != 0 { //1 is cloudbe
			msg, err = cloudbe.RecvMessage(0)
			if err != nil {
				log.Fatalln(err)
			}

			//we don't need peer broker id
			_, msg = unwrap(msg)
		}

		if len(msg) > 0 {
			toBroker := false

			for _, peer := range peers {
				if peer == msg[0] {
					toBroker = true
					break
				}
			}

			if toBroker {
				cloudfe.SendMessage(msg)
			} else {
				localfe.SendMessage(msg)
			}
		}

		//procces state fe and monitor
		// 2 is statefe
		if sockets[2].Events&zmq4.POLLIN != 0 {
			msg, err = statefe.RecvMessage(0)
			if err != nil {
				log.Fatalln(err)
			}

			//peer
			_, msg = unwrap(msg)

			status, _ := unwrap(msg)
			cloudCapacity, err = strconv.Atoi(status)
		}

		//3 is monitor
		if sockets[3].Events&zmq4.POLLIN != 0 {
			status, err := monitor.Recv(0)
			if err != nil {
				log.Fatalln(err)
			}

			fmt.Println(status)
		}

		//route locally as many client requests as handled
		for localCapacity+cloudCapacity > 0 {
			if localCapacity > 0 {
				sockets, err = secondary2.PollAll(0)
				if err != nil {
					log.Fatalln(err)
				}
			} else {
				sockets, err = secondary1.PollAll(0)
				if err != nil {
					log.Fatalln(err)
				}
			}

			//0 is localfe
			if sockets[0].Events&zmq4.POLLIN != 0 {
				msg, err = localfe.RecvMessage(0)
				if err != nil {
					log.Fatalln(err)
				}
			} else if len(sockets) > 1 && sockets[1].Events&zmq4.POLLIN != 0 { //1 is cloudfe
				msg, err = cloudfe.RecvMessage(0)
				fmt.Println("cloudfe")
				if err != nil {
					log.Fatalln(err)
				}
			} else {
				break
			}

			if localCapacity > 0 {
				localbe.SendMessage(workerQueue[0], "", msg)
				workerQueue = workerQueue[1:]
				localCapacity--
			} else {
				randomPeer := rand.Intn(len(peers)-2) + 2
				cloudbe.SendMessage(peers[randomPeer], "", msg)
			}
		}

		//broadcast capacity messages to other peers in case of capacity change
		if localCapacity != previousCapacity {
			statebe.SendMessage(self, "", localCapacity)
		}

	}

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
