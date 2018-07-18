package servers

import (
	"github.com/pebbe/zmq4"
	"log"
	"time"
)

func EnvPub() {
	publisher, err := zmq4.NewSocket(zmq4.PUB)
	if err != nil {
		log.Fatalln(err)
	}

	defer publisher.Close()

	err = publisher.Bind("tcp://*:5563")
	if err != nil {
		log.Fatalln(err)
	}

	for {
		_, err = publisher.Send("A", zmq4.SNDMORE)
		if err != nil {
			log.Fatalln(err)
		}

		_, err = publisher.Send("I don't want to see it", 0)
		if err != nil {
			log.Fatalln(err)
		}

		_, err = publisher.Send("B", zmq4.SNDMORE)
		if err != nil {
			log.Fatalln(err)
		}

		_, err = publisher.Send("Better", 0)
		if err != nil {
			log.Fatalln(err)
		}

		time.Sleep(time.Second)

	}
}
