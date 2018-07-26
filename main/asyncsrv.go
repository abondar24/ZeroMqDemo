package main

import (
	"fmt"
	"github.com/pebbe/zmq4"
	"log"
	"math/rand"
	"sync"
	"time"
)

//sends requests once per second
//collects respomses upon arrival
func client() {

	var mu sync.Mutex

	client, err := zmq4.NewSocket(zmq4.DEALER)
	if err != nil {
		log.Fatalln(err)
	}

	defer client.Close()

	setId(client)

	client.Connect("tcp://localhost:5570")

	go func() {
		for reqNum := 1; true; reqNum++ {
			time.Sleep(time.Second)
			mu.Lock()
			client.SendMessage(fmt.Sprintf("request #%d", reqNum))
			mu.Unlock()
		}
	}()

	for {
		time.Sleep(10 * time.Millisecond)
		mu.Lock()

		msg, err := client.RecvMessage(zmq4.DONTWAIT)
		if err != nil {
			log.Println(err)
		} else {
			id, err := client.GetIdentity()
			if err != nil {
				log.Println(err)
			}
			fmt.Println(msg[0], id)
		}
		mu.Unlock()
	}

}

func server() {
	frontend, err := zmq4.NewSocket(zmq4.ROUTER)
	if err != nil {
		log.Fatalln(err)
	}

	defer frontend.Close()
	frontend.Bind("tcp://*:5570")

	backend, err := zmq4.NewSocket(zmq4.DEALER)
	if err != nil {
		log.Fatalln(err)
	}

	defer backend.Close()
	backend.Bind("inproc://backend")

	for i := 0; i < 5; i++ {
		go worker()
	}

	err = zmq4.Proxy(frontend, backend, nil)
	if err != nil {
		log.Fatalln(err)
	}

}

// each worker works on one request per time
func worker() {
	worker, err := zmq4.NewSocket(zmq4.DEALER)
	if err != nil {
		log.Fatalln(err)
	}

	defer worker.Close()
	worker.Connect("inproc://backend")

	for {
		msg, err := worker.RecvMessage(0)
		if err != nil {
			log.Println(err)
		}

		id, content := pop(msg)

		replies := rand.Intn(5)
		for reply := 0; reply < replies; reply++ {
			time.Sleep(time.Duration(rand.Intn(1000)+1) * time.Millisecond)
			worker.SendMessage(id, content)
		}
	}
}

func setId(soc *zmq4.Socket) {
	id := fmt.Sprintf("%04X-%04X", rand.Intn(0x10000), rand.Intn(0x10000))
	soc.SetIdentity(id)
}

func pop(msg []string) (head, tail []string) {
	if msg[1] == "" {
		head = msg[:2]
		tail = msg[2:]
	} else {
		head = msg[:1]
		tail = msg[1:]
	}

	return
}

func AsyncServer() {
	rand.Seed(time.Now().UnixNano())

	go client()
	go client()
	go client()
	go server()

	time.Sleep(5 * time.Second)
}
