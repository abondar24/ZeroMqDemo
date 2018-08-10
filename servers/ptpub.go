package servers

import (
	"fmt"
	"github.com/pebbe/zmq4"
	"log"
	"math/rand"
	"time"
)

func PathologicalPublisher(cacheEnable bool) {

	publisher, err := zmq4.NewSocket(zmq4.PUB)
	if err != nil {
		log.Fatalln(err)
	}

	if cacheEnable {
		publisher.Connect("tcp://localhost:5557")
	} else {
		publisher.Bind("tcp://*:5556")
	}

	time.Sleep(time.Second)

	for topicNum := 0; topicNum < 1000; topicNum++ {
		_, err := publisher.SendMessage(fmt.Sprintf("%03d", topicNum), "SOS")
		if err != nil {
			fmt.Println(err)
		}

	}

	rand.Seed(time.Now().UnixNano())
	for {
		time.Sleep(time.Second)
		_, err := publisher.SendMessage(fmt.Sprintf("%03d", rand.Intn(1000)), "Target gone")
		if err != nil {
			fmt.Println(err)
		}
	}
}
