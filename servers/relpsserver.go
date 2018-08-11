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
	publisher.Bind("tcp://*:5556")
	time.Sleep(200 * time.Millisecond)

	kvmap := make(map[string]*kvmessage.KVmsg)
	rand.Seed(time.Now().UnixNano())

	sequence := int64(1)
	for ; true; sequence++ {
		kvmsg := kvmessage.NewKVMessage(sequence)
		kvmsg.SetKey(fmt.Sprint(rand.Intn(10000)))
		kvmsg.SetBody(fmt.Sprint(rand.Intn(1000000)))
		err := kvmsg.SendKVmsg(publisher)
		kvmsg.StoreMsg(kvmap)
		if err != nil {
			log.Fatalln(err)
		}
	}

	fmt.Printf("Interrupted\n%d messages out\n", sequence)
}
