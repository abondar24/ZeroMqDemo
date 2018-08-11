package clients

import (
	"fmt"
	"github.com/pebbe/zmq4"
	"github.com/pebbe/zmq4/examples/kvsimple"
	"log"
)

func ReliablePubSubClient() {
	updates, err := zmq4.NewSocket(zmq4.SUB)
	if err != nil {
		log.Fatalln(err)
	}

	updates.SetSubscribe("")
	updates.Connect("tcp://localhost:5556")

	kvmap := make(map[string]*kvsimple.Kvmsg)

	sequence := int64(0)
	for ; true; sequence++ {
		kvmsg, err := kvsimple.RecvKvmsg(updates)
		if err != nil {
			log.Fatalln(err)
		}

		kvmsg.Store(kvmap)
		seq, err := kvmsg.GetSequence()
		if err != nil {
			log.Fatalln(err)
		}

		key, err := kvmsg.GetKey()
		if err != nil {
			log.Fatalln(err)
		}

		fmt.Println(key, ":", seq)
	}
	fmt.Printf("Interrupted\n%d messages in\n", sequence)
}
