package clients

import (
	"fmt"
	"github.com/pebbe/zmq4"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func Interrupt() {
	client, err := zmq4.NewSocket(zmq4.REQ)
	if err != nil {
		log.Fatalln(err)
	}

	defer client.Close()

	err = client.Connect("tcp://localhost:5555")
	if err != nil {
		log.Fatalln(err)
	}

	chSignal := make(chan os.Signal, 1)
	signal.Notify(chSignal, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM)

LOOP:
	for {
		_, err = client.Send("HELLO", 0)
		if err != nil {
			log.Println(err)
		}

		fmt.Println("Sent: Hello")

		reply, err := client.Recv(0)
		if err != nil {
			if zmq4.AsErrno(err) == zmq4.Errno(syscall.EINTR) {
				log.Println("Client Recv:", err)
				break

			} else {
				log.Panicln(err)
			}
		}

		fmt.Println("Received:", reply)
		time.Sleep(time.Second)

		select {
		case signal := <-chSignal:
			log.Println("Signal:", signal)
			break LOOP
		default:

		}
	}
}
