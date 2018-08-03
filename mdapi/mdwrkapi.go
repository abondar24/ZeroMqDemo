package mdapi

import (
	"github.com/pebbe/zmq4"
	"log"
	"runtime"
	"time"
)

const (
	HeartbeatLiveness = 3
)

type MdWorker struct {
	broker  string
	service string
	worker  *zmq4.Socket
	poller  *zmq4.Poller
	verbose bool

	heartbeatAt time.Time
	liveness    int
	heartbeat   time.Duration
	reconnect   time.Duration

	expectReply bool
	replyTo     string
}

func (mdwork *MdWorker) SendToBroker(command string, option string, msg []string) (err error) {
	n := 3
	if option != "" {
		n++
	}

	m := make([]string, n, n+len(msg))
	m = append(m, msg...)

	if option != "" {
		m[3] = option
	}

	m[2] = command
	m[1] = MdWorkerVer
	m[0] = ""

	if mdwork.verbose {
		log.Printf("Sending %s to broker %q\n", MdCommands[command], m)
	}

	_, err = mdwork.worker.SendMessage(m)
	return
}

func (mdwork *MdWorker) ConnectToBroker() (err error) {
	if mdwork.worker != nil {
		mdwork.worker.Close()
		mdwork.worker = nil
	}

	mdwork.worker, err = zmq4.NewSocket(zmq4.DEALER)
	if err != nil {
		return
	}

	err = mdwork.worker.Connect(mdwork.broker)
	if mdwork.verbose {
		log.Printf("Connecting to broker at %s...\n", mdwork.broker)
	}

	mdwork.poller = zmq4.NewPoller()
	mdwork.poller.Add(mdwork.worker, zmq4.POLLIN)

	err = mdwork.SendToBroker(MdReady, mdwork.service, []string{})

	mdwork.liveness = HeartbeatLiveness
	mdwork.heartbeatAt = time.Now().Add(mdwork.heartbeat)

	return
}

func NewWorker(broker, service string, verbose bool) (mdworker *MdWorker, err error) {
	mdworker = &MdWorker{
		broker:    broker,
		service:   service,
		verbose:   verbose,
		heartbeat: 2500 * time.Millisecond,
		reconnect: 2500 * time.Millisecond,
	}

	err = mdworker.ConnectToBroker()
	runtime.SetFinalizer(mdworker, (*MdWorker).Close)

	return
}

func (mdwork *MdWorker) Close() {
	if mdwork.worker != nil {
		mdwork.worker.Close()
		mdwork.worker = nil
	}
}

func (mdwork *MdWorker) SetHeartbeat(heartbeat time.Duration) {
	mdwork.heartbeat = heartbeat
}

func (mdwork *MdWorker) SetReconnect(reconnect time.Duration) {
	mdwork.reconnect = reconnect
}

func (mdwork *MdWorker) SendReply(reply []string) (msg []string, err error) {
	if len(reply) == 0 && mdwork.expectReply {
		log.Fatalln("No reply expected")
	}

	if len(reply) > 0 {
		if mdwork.replyTo == "" {
			log.Fatalln("mdwork.replyTo==\"\"")
		}
		m := make([]string, 2, 2+len(reply))
		m = append(m, reply...)
		m[0] = mdwork.replyTo
		m[1] = ""
		err = mdwork.SendToBroker(MdReply, "", m)
	}

	mdwork.expectReply = true

	for {
		var polled []zmq4.Polled
		polled, err = mdwork.poller.PollAll(mdwork.heartbeat)
		if err != nil {
			log.Fatalln(err)
		}

		if len(polled) > 0 {
			msg, err = mdwork.worker.RecvMessage(0)
			if err != nil {
				log.Fatalln(err)
			}

			if mdwork.verbose {
				log.Printf("Received message from broker: %q\n", msg)
			}
			mdwork.liveness = HeartbeatLiveness

			if len(msg) < 3 {
				log.Fatalln("len(msg) < 3")
			}

			if msg[0] != "" {
				log.Fatalln("msg[0] !=\"\"")
			}

			if msg[1] != MdWorkerVer {
				log.Fatalln("msg[1] != MdWorkerVer")
			}

			command := msg[2]
			msg = msg[3:]

			switch command {
			case MdRequest:
				mdwork.replyTo, msg = unwrap(msg)

			case MdHearbeart:
			case MdDisconnect:
				mdwork.ConnectToBroker()
			default:
				log.Printf("Invalid input message %q\n", msg)
			}
		} else {
			mdwork.liveness--
			if mdwork.liveness == 0 {
				if mdwork.verbose {
					log.Println("Disconnected from broker - retrying...")
				}
				time.Sleep(mdwork.reconnect)
				mdwork.ConnectToBroker()
			}
		}

		if time.Now().After(mdwork.heartbeatAt) {
			mdwork.SendToBroker(MdHearbeart, "", []string{})
			mdwork.heartbeatAt = time.Now().Add(mdwork.heartbeat)
		}
	}

	return
}
