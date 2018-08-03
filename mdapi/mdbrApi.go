package mdapi

import (
	"fmt"
	"github.com/pebbe/zmq4"
	"log"
	"runtime"
	"time"
)

type Broker struct {
	Socket      *zmq4.Socket
	Verbose     bool
	endpoint    string
	services    map[string]*Service
	workers     map[string]*Worker
	Waiting     []*Worker
	HeartbeatAt time.Time
}

type Service struct {
	broker   *Broker
	name     string
	requests [][]string
	waiting  []*Worker
}

type Worker struct {
	broker  *Broker
	idStr   string
	id      string
	service *Service
	expiry  time.Time
}

func NewBroker(verbose bool) (broker *Broker, err error) {
	broker = &Broker{
		Verbose:     verbose,
		services:    make(map[string]*Service),
		workers:     make(map[string]*Worker),
		Waiting:     make([]*Worker, 0),
		HeartbeatAt: time.Now().Add(HeartbeatInterval),
	}

	broker.Socket, err = zmq4.NewSocket(zmq4.ROUTER)

	broker.Socket.SetRcvhwm(500000)

	runtime.SetFinalizer(broker, (*Broker).Close)
	return
}

func (broker *Broker) Close() (err error) {
	if broker.Socket != nil {
		err = broker.Socket.Close()
		broker.Socket = nil
	}

	return
}

func (broker *Broker) Bind(endpoint string) (err error) {
	err = broker.Socket.Bind(endpoint)
	if err != nil {
		log.Println("MDP broker/0.1.0 failed to bind at", endpoint)
		return
	}
	log.Println("MDP broker/0.1.0 is active at", endpoint)
	return
}

func (broker *Broker) WorkerMsg(sender string, msg []string) {
	if len(msg) == 0 {
		log.Fatalln("len(msg) == 0")
	}

	command, msg := popStr(msg)
	idString := fmt.Sprintf("%q", sender)
	_, WorkerReady := broker.workers[idString]
	worker := broker.WorkerRequire(sender)

	switch command {
	case MdReady:
		if WorkerReady {
			worker.Delete(true)
		} else if len(sender) >= 4 && sender[:4] == "mmi." {
			worker.Delete(true)
		} else {
			worker.service = broker.ServiceRequire(msg[0])
			worker.Waiting()
		}
	case MdReply:
		if WorkerReady {
			client, msg := unwrap(msg)
			broker.Socket.SendMessage(client, "", MdClientVer, worker.service.name, msg)
			worker.Waiting()
		} else {
			worker.Delete(true)
		}
	case MdHearbeart:
		if WorkerReady {
			worker.expiry = time.Now().Add(HeartbeatExpiry)
		} else {
			worker.Delete(false)
		}
	case MdDisconnect:
		worker.Delete(false)
	default:
		log.Printf("Invalid input message %q\n", msg)
	}
}

func (broker *Broker) ClientMsg(sender string, msg []string) {
	if len(msg) < 2 {
		log.Fatalln("len(msg) < 2")
	}

	serviceFrame, msg := popStr(msg)
	service := broker.ServiceRequire(serviceFrame)

	m := []string{sender, ""}
	msg = append(m, msg...)

	if len(serviceFrame) >= 4 && serviceFrame[:4] == "mmi." {
		var returnCode string
		if serviceFrame == "mmi.service" {
			name := msg[len(msg)-1]
			service, ok := broker.services[name]
			if ok && len(service.waiting) > 0 {
				returnCode = "200"
			} else {
				returnCode = "404"
			}
		} else {
			returnCode = "501"
		}

		msg[len(msg)-1] = returnCode

		client, msg := unwrap(msg)
		broker.Socket.SendMessage(client, "", MdClientVer, serviceFrame, msg)
	} else {
		service.Dispatch(msg)
	}
}

func (broker *Broker) Purge() {
	now := time.Now()
	for len(broker.Waiting) > 0 {
		if broker.Waiting[0].expiry.After(now) {
			break
		}
		if broker.Verbose {
			log.Println("Deleting expired worker:", broker.Waiting[0].idStr)
		}
		broker.Waiting[0].Delete(false)
	}
}

func (broker *Broker) ServiceRequire(serviceFrame string) (service *Service) {
	name := serviceFrame
	service, ok := broker.services[name]
	if !ok {
		service = &Service{
			broker:   broker,
			name:     name,
			requests: make([][]string, 0),
			waiting:  make([]*Worker, 0),
		}
		broker.services[name] = service
		if broker.Verbose {
			log.Println("Added service:", name)
		}
	}

	return
}

func (broker *Broker) WorkerRequire(id string) (worker *Worker) {
	idStr := fmt.Sprintf("%q", id)
	worker, ok := broker.workers[idStr]
	if !ok {
		worker = &Worker{
			broker: broker,
			idStr:  idStr,
			id:     id,
		}
		broker.workers[idStr] = worker
		if broker.Verbose {
			log.Printf("Registering new worker: %s\n", idStr)
		}
	}
	return
}

// send requests to Waiting workers
func (service *Service) Dispatch(msg []string) {
	if len(msg) > 0 {
		service.requests = append(service.requests, msg)
	}

	service.broker.Purge()
	for len(service.waiting) > 0 && len(service.requests) > 0 {
		var worker *Worker

		worker, service.waiting = popWorker(service.waiting)
		service.broker.Waiting = delWorker(service.broker.Waiting, worker)
		msg, service.requests = popMsg(service.requests)
		worker.Send(MdRequest, "", msg)
	}
}

//delete current worker
func (worker *Worker) Delete(disconnect bool) {
	if disconnect {
		worker.Send(MdDisconnect, "", []string{})
	}

	if worker.service != nil {
		worker.service.waiting = delWorker(worker.service.waiting, worker)
	}

	worker.broker.Waiting = delWorker(worker.broker.Waiting, worker)
	delete(worker.broker.workers, worker.idStr)
}

//format and send command to worker
func (worker *Worker) Send(command, option string, msg []string) (err error) {
	n := 4
	if option != "" {
		n++
	}

	m := make([]string, n, n+len(msg))
	m = append(m, msg...)

	if option != "" {
		m[4] = option
	}
	m[3] = command
	m[2] = MdWorkerVer
	m[1] = ""
	m[0] = worker.id

	if worker.broker.Verbose {
		log.Printf("Sending %s to worker %q\n", MdCommands[command], m)
	}

	_, err = worker.broker.Socket.SendMessage(m)
	return
}

func (worker *Worker) Waiting() {
	worker.broker.Waiting = append(worker.broker.Waiting, worker)
	worker.service.waiting = append(worker.service.waiting, worker)
	worker.expiry = time.Now().Add(HeartbeatExpiry)
	worker.service.Dispatch([]string{})
}
