package brokers

import (
	"github.com/pebbe/zmq4"
)

type Broker struct {
	frontend    *zmq4.Socket
	backend     *zmq4.Socket
	workerQueue []string
	reactor     *zmq4.Reactor
}

func handleFrontend(broker *Broker) error {
	msg, err := broker.frontend.RecvMessage(0)
	if err != nil {
		return err
	}

	broker.backend.SendMessage(broker.workerQueue[0], "", msg)
	broker.workerQueue = broker.workerQueue[1:]

	if len(broker.workerQueue) == 0 {
		broker.reactor.RemoveSocket(broker.frontend)
	}
	return nil
}

func handleBackend(broker *Broker) error {
	msg, err := broker.backend.RecvMessage(0)
	if err != nil {
		return err
	}

	id, msg := unwrap(msg)
	broker.workerQueue = append(broker.workerQueue, id)

	//enable reader on frontend
	if len(broker.workerQueue) == 1 {
		broker.reactor.AddSocket(broker.frontend, zmq4.POLLIN,
			func(e zmq4.State) error {
				return handleFrontend(broker)
			})
	}

	if msg[0] != WorkerReady {
		broker.frontend.SendMessage(msg)
	}

	return nil
}

func LoadBalacningReactorBroker() {

	broker := &Broker{}

	broker.frontend, _ = zmq4.NewSocket(zmq4.ROUTER)
	defer broker.frontend.Close()
	broker.frontend.Bind("ipc://frontend.ipc")

	broker.backend, _ = zmq4.NewSocket(zmq4.ROUTER)
	defer broker.backend.Close()
	broker.backend.Bind("ipc://backend.ipc")

	for clientsNum := 0; clientsNum < NumberOfClients; clientsNum++ {
		go clientTask()
	}

	for workerNum := 0; workerNum < NumberOfWorkers; workerNum++ {
		go workerTask()
	}

	broker.workerQueue = make([]string, 0, 10)

	broker.reactor = zmq4.NewReactor()
	broker.reactor.AddSocket(broker.backend, zmq4.POLLIN, func(e zmq4.State) error {
		return handleBackend(broker)
	})
	broker.reactor.Run(-1)
}
