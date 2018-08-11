package servers

import (
	"fmt"
	"github.com/abondar24/ZeroMqDemo/rpsapi"
	"github.com/pebbe/zmq4"
	"log"
	"strings"
	"time"
)

func ReliablePubSubServer() {

	//state request
	snapshot, err := zmq4.NewSocket(zmq4.ROUTER)
	if err != nil {
		log.Fatalln(err)
	}
	snapshot.Bind("tcp://*:5556")

	//updates
	publisher, err := zmq4.NewSocket(zmq4.PUB)
	if err != nil {
		log.Fatalln(err)
	}
	publisher.Bind("tcp://*:5557")

	//state updates
	collector, err := zmq4.NewSocket(zmq4.PULL)
	if err != nil {
		log.Fatalln(err)
	}
	collector.Bind("tcp://*:5558")

	kvmap := make(map[string]*rpsapi.KVmsg)
	sequence := int64(0)

	poller := zmq4.NewPoller()
	poller.Add(collector, zmq4.POLLIN)
	poller.Add(snapshot, zmq4.POLLIN)

LOOP:
	for {
		polled, err := poller.Poll(1000 * time.Millisecond)
		if err != nil {
			break
		}

		for _, item := range polled {
			switch socket := item.Socket; socket {
			case collector:
				kvmsg, err := rpsapi.RecvKVmsg(collector)
				if err != nil {
					break LOOP
				}
				sequence++
				kvmsg.SetSequence(sequence)
				kvmsg.SendKVmsg(publisher)
				kvmsg.StoreMsg(kvmap)
				fmt.Println("Publishing state", sequence)
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

				subtree := msg[2]

				for _, kvmsg := range kvmap {
					if key, _ := kvmsg.GetKey(); strings.HasPrefix(key, subtree) {
						snapshot.Send(id, zmq4.SNDMORE)
						kvmsg.SendKVmsg(snapshot)
					}
				}

				fmt.Printf("Sending state snaphot=%d\n", sequence)
				snapshot.Send(id, zmq4.SNDMORE)
				kvmsg := rpsapi.NewKVMessage(sequence)
				kvmsg.SetKey("KTHXBAI")
				kvmsg.SetBody("")
				kvmsg.SendKVmsg(snapshot)
			}
		}
	}

	fmt.Printf("Interrupted\n%d messages out\n", sequence)
}
