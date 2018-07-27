package servers

import (
	"fmt"
	"github.com/pebbe/zmq4"
	"log"
	"math/rand"
	"time"
)

func LazyServer() {
	rand.Seed(time.Now().UnixNano())

	server, err := zmq4.NewSocket(zmq4.REP)
	if err != nil {
		log.Fatalln(err)
	}

	defer server.Close()
	server.Bind("tcp://*:5555")

	for cycles := 0; true; {
		request, err := server.RecvMessage(0)
		if err != nil {
			log.Fatalln(err)
		}
		cycles++

		//simulate issues
		if cycles > 4 && rand.Intn(3) == 0 {
			fmt.Println("Simulating crash")
			break
		} else if cycles > 4 && rand.Intn(3) == 0 {
			fmt.Println("Simulate CPU overload")
			time.Sleep(2 * time.Second)
		}

		fmt.Printf("Normal request (%s)\n", request)
		time.Sleep(time.Second)
		server.SendMessage(request)

	}
}
