package servers

import (
	"log"
	"github.com/pebbe/zmq4"
	"math/rand"
	"time"
	"fmt"
)

func WuServer(){


	socket, err := zmq4.NewSocket(zmq4.PUB)
	if err!=nil{
		log.Fatal(err)
	}

	defer  socket.Close()

	socket.Bind("tcp://*:5556")
	socket.Bind("ipc://weather.ipc")

	rand.Seed(time.Now().UnixNano())

	for {
		zipcode := rand.Intn(1000000)
		temperature := rand.Intn(215) - 80
		humidity := rand.Intn(50) + 10

		msg := fmt.Sprintf("%d %d %d",zipcode,temperature,humidity)

		socket.Send(msg,0)
	}
}
