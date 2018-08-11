package servers

import (
	"github.com/abondar24/ZeroMqDemo/kvmessage"
	"github.com/pebbe/zmq4"
	"log"
	"math/rand"
	"time"

	"fmt"
)

func ReliablePubSubServer() {
	publisher, err := zmq4.NewSocket(zmq4.PUB)
	if err != nil {
		log.Fatalln(err)
	}
	publisher.Bind("tcp://*:5557")

	rand.Seed(time.Now().UnixNano())
	sequence := int64(1)

	updates, err := zmq4.NewSocket(zmq4.PAIR)
	if err != nil {
		log.Fatalln(err)
	}
	updates.Bind("inproc://pipe")
	go stateManager()
	updates.RecvMessage(0)

	for {
		sequence++
		kvmsg := kvmessage.NewKVMessage(sequence)
		kvmsg.SetKey(fmt.Sprint(rand.Intn(10000)))
		kvmsg.SetBody(fmt.Sprint(rand.Intn(1000000)))
		err := kvmsg.SendKVmsg(publisher)
		if err != nil {
			log.Fatalln(err)
		}

		err = kvmsg.SendKVmsg(updates)
		if err != nil {
			log.Fatalln(err)
		}
	}

	fmt.Printf("Interrupted\n%d messages out\n", sequence)
}

func stateManager() {
	kvmap := make(map[string]*kvmessage.KVmsg)

	pipe, err := zmq4.NewSocket(zmq4.PAIR)
	if err != nil {
		log.Fatalln(err)
	}

	pipe.Connect("inproc://pipe")
	pipe.SendMessage("READY")

	snapshot, err := zmq4.NewSocket(zmq4.ROUTER)
	if err != nil {
		log.Fatalln(err)
	}
	snapshot.Bind("tcp://*:5556")

	poller := zmq4.NewPoller()
	poller.Add(pipe, zmq4.POLLIN)
	poller.Add(snapshot, zmq4.POLLIN)
	sequence := int64(0)

LOOP:
	for {
		polled, err := poller.Poll(-1)
		if err != nil {
			break
		}

		for _, item := range polled {
			switch socket := item.Socket; socket {
			case pipe:
				kvmsg, err := kvmessage.RecvKVmsg(pipe)
				if err != nil {
					break LOOP
				}
				sequence, err = kvmsg.GetSequence()
				if err != nil {
					log.Fatalln(err)
				}

				kvmsg.StoreMsg(kvmap)

			case snapshot:
				msg, err := snapshot.RecvMessage(0)
				if err != nil {
					break LOOP
				}

				id := msg[0]
				request := msg[1]
				if request != "ICANHAZ?" {
					fmt.Println("Bad request")
					break LOOP
				}

				for _, kvmsg := range kvmap {
					snapshot.Send(id, zmq4.SNDMORE)
					kvmsg.SendKVmsg(snapshot)
				}

				//time for client to get msg
				time.Sleep(100 * time.Millisecond)

				fmt.Printf("Sending state snaphot=%d\n", sequence)
				snapshot.Send(id, zmq4.SNDMORE)
				kvmsg := kvmessage.NewKVMessage(sequence)
				kvmsg.SetKey("KTHXBAI")
				kvmsg.SetBody("")
				kvmsg.SendKVmsg(snapshot)
			}
		}
	}
}
