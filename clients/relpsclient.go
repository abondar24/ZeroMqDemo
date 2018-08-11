package clients

import (
	"fmt"
	"github.com/abondar24/ZeroMqDemo/kvmessage"
	"github.com/pebbe/zmq4"
	"log"
	"time"
)

func ReliablePubSubClient() {

	snapshot, err := zmq4.NewSocket(zmq4.DEALER)
	if err != nil {
		log.Fatalln(err)
	}
	snapshot.Connect("tcp://localhost:5556")

	subscriber, err := zmq4.NewSocket(zmq4.SUB)
	if err != nil {
		log.Fatalln(err)
	}

	// or messages between snapshot and next are lost
	subscriber.SetRcvhwm(100000)
	subscriber.SetSubscribe("")
	subscriber.Connect("tcp://localhost:5557")

	time.Sleep(time.Second)

	kvmap := make(map[string]*kvmessage.KVmsg)

	sequence := int64(0)
	snapshot.SendMessage("ICANHAZ?")
	for {
		kvmsg, err := kvmessage.RecvKVmsg(snapshot)
		if err != nil {
			log.Fatalln(err)
		}

		if key, _ := kvmsg.GetKey(); key == "KTHXBAI" {
			sequence, err := kvmsg.GetSequence()
			if err != nil {
				log.Fatalln(err)
			}

			fmt.Printf("Received snapshot=%d\n", sequence)
			break
		}
		kvmsg.StoreMsg(kvmap)
	}
	snapshot.Close()

	first := true

	for {
		kvmsg, err := kvmessage.RecvKVmsg(subscriber)
		if err != nil {
			log.Fatalln(err)
		}

		if seq, _ := kvmsg.GetSequence(); seq > sequence {
			sequence, err := kvmsg.GetSequence()
			if err != nil {
				log.Fatalln(err)
			}

			kvmsg.StoreMsg(kvmap)
			if first {
				//show first update after snapshot
				first = false
				fmt.Println("Next:", sequence)
			}
		}
	}
}
