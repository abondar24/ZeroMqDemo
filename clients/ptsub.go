package clients

import (
	"fmt"
	"github.com/pebbe/zmq4"
	"log"
	"math/rand"
	"time"
)

func PathologicalSubscriber(cacheEnable bool) {
	subscriber, err := zmq4.NewSocket(zmq4.SUB)
	if err != nil {
		log.Fatalln(err)
	}

	if cacheEnable {
		subscriber.Connect("tcp://localhost:5558")
	} else {
		subscriber.Connect("tcp://localhost:5556")
	}

	rand.Seed(time.Now().UnixNano())
	subscription := fmt.Sprintf("%03d", rand.Intn(1000))
	subscriber.SetSubscribe(subscription)

	for {
		msg, err := subscriber.RecvMessage(0)
		if err != nil {
			log.Fatalln(err)
		}

		topic := msg[0]
		data := msg[1]
		if topic != subscription {
			log.Fatalln("topic != subscription")
		}

		fmt.Println(topic)
		fmt.Println(data)
	}
}
