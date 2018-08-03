package brokers

import (
	"github.com/abondar24/ZeroMqDemo/mdapi"
	"github.com/pebbe/zmq4"
	"log"
	"time"
)

func MajordomoBroker(verbose bool) {
	broker, err := mdapi.NewBroker(verbose)
	if err != nil {
		log.Fatalln(err)
	}

	broker.Bind("tcp://*:5555")

	poller := zmq4.NewPoller()
	poller.Add(broker.Socket, zmq4.POLLIN)

	for {
		polled, err := poller.Poll(mdapi.HeartbeatInterval)
		if err != nil {
			log.Fatalln(err)
		}

		if len(polled) > 0 {
			msg, err := broker.Socket.RecvMessage(0)
			if err != nil {
				log.Fatalln(err)
			}

			if broker.Verbose {
				log.Printf("Received message: %q\n", msg)
			}

			sender, msg := popStr(msg)
			_, msg = popStr(msg)
			header, msg := popStr(msg)

			switch header {
			case mdapi.MdClientVer:
				broker.ClientMsg(sender, msg)
			case mdapi.MdWorkerVer:
				broker.WorkerMsg(sender, msg)
			default:
				log.Printf("Invalid message", msg)
			}
		}

		if time.Now().After(broker.HeartbeatAt) {
			broker.Purge()
			for _, worker := range broker.Waiting {
				worker.Send(mdapi.MdHearbeart, "", []string{})
			}
			broker.HeartbeatAt = time.Now().Add(mdapi.HeartbeatInterval)
		}

	}
	log.Println("Interrupt received,shutting down...")
}

func popStr(ss []string) (s string, ss2 []string) {
	s = ss[0]
	ss2 = ss[1:]
	return
}
