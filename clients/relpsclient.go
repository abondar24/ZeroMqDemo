package clients

import (
	"fmt"
	"github.com/abondar24/ZeroMqDemo/rpsapi"
	"github.com/pebbe/zmq4"
	"log"
	"math/rand"
	"time"
)

const Subtree = "/client/"

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
	subscriber.Connect("tcp://localhost:5557")

	publisher, err := zmq4.NewSocket(zmq4.PUSH)
	if err != nil {
		log.Fatalln(err)
	}
	publisher.Connect("tcp://localhost:5558")

	rand.Seed(time.Now().UnixNano())
	kvmap := make(map[string]*rpsapi.KVmsg)

	sequence := int64(0)
	snapshot.SendMessage("ICANHAZ?", Subtree)
	for {
		kvmsg, err := rpsapi.RecvKVmsg(snapshot)
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

	poller := zmq4.NewPoller()
	poller.Add(subscriber, zmq4.POLLIN)
	alarm := time.Now().Add(1000 * time.Millisecond)

	for {
		tickless := alarm.Sub(time.Now())
		if tickless < 0 {
			tickless = 0
		}

		polled, err := poller.Poll(tickless)
		if err != nil {
			log.Fatalln(err)
		}

		if len(polled) == 1 {
			kvmsg, err := rpsapi.RecvKVmsg(subscriber)
			if err != nil {
				log.Fatalln(err)
			}

			if seq, _ := kvmsg.GetSequence(); seq > sequence {
				sequence = seq
				kvmsg.StoreMsg(kvmap)
				fmt.Println("Received update =", sequence)
			}
		}

		//timed out, send random kvmsg
		if time.Now().After(alarm) {
			kvmsg := rpsapi.NewKVMessage(0)
			kvmsg.SetKey(fmt.Sprintf("%s%d", Subtree, rand.Intn(10000)))
			kvmsg.SetBody(fmt.Sprint(rand.Intn(1000000)))
			kvmsg.SendKVmsg(publisher)
			alarm = time.Now().Add(1000 * time.Millisecond)
		}
	}

	fmt.Printf("Interrupted\n%d messages in\n", sequence)
}
