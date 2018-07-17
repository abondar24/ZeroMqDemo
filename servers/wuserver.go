package servers

import (
	"fmt"
	"github.com/pebbe/zmq4"
	"log"
	"math/rand"
	"time"
)

func WuServer() {

	socket, err := zmq4.NewSocket(zmq4.PUB)
	if err != nil {
		log.Fatalln(err)
	}

	defer socket.Close()

	if socket.Bind("tcp://*:5556") != nil {
		log.Fatalln(err)
	}

	if socket.Bind("ipc://weather.ipc") != nil {
		log.Fatalln(err)
	}

	rand.Seed(time.Now().UnixNano())

	for {
		zipcode := rand.Intn(1000000)
		temperature := rand.Intn(215) - 80
		humidity := rand.Intn(50) + 10

		msg := fmt.Sprintf("%d %d %d", zipcode, temperature, humidity)

		_, err = socket.Send(msg, 0)
		if err != nil {
			log.Println(err)
		}
	}
}
