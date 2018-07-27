package clients

import (
	"fmt"
	"github.com/pebbe/zmq4"
	"log"
	"strconv"
	"time"
)

const (
	RequestTimeout = 2500 * time.Millisecond
	RequestRetries = 3
	ServerEndpoint = "tcp://localhost:5555"
)

func LazyClient() {
	fmt.Println("Connecting to server")
	client, err := zmq4.NewSocket(zmq4.REQ)
	if err != nil {
		log.Fatalln(err)
	}
	client.Connect(ServerEndpoint)

	poller := zmq4.NewPoller()
	poller.Add(client, zmq4.POLLIN)

	sequence := 0
	retriesLeft := RequestRetries

	for retriesLeft > 0 {
		sequence++
		client.SendMessage(sequence)

		for expectReply := true; expectReply; {
			sockets, err := poller.Poll(RequestTimeout)
			if err != nil {
				log.Fatalln(err)
			}

			if len(sockets) > 0 {
				reply, err := client.RecvMessage(0)
				if err != nil {
					log.Fatalln(err)
				}

				seq, err := strconv.Atoi(reply[0])
				if err != nil {
					log.Fatalln(err)
				}

				if seq == sequence {
					fmt.Printf("Server replied OK (%s)\n", reply[0])
					retriesLeft = RequestRetries
					expectReply = false
				} else {
					fmt.Printf("Malformed reply from server: %s\n", reply)
				}
			} else {
				retriesLeft--
				if retriesLeft == 0 {
					fmt.Println("Server is offline")
					break
				} else {
					fmt.Println("No server response, retrying...")

					client.Close()
					client, err = zmq4.NewSocket(zmq4.REQ)
					if err != nil {
						log.Fatalln(err)
					}
					client.Connect(ServerEndpoint)
					poller = zmq4.NewPoller()
					poller.Add(client, zmq4.POLLIN)
					client.SendMessage(sequence)
				}
			}
		}
	}
	client.Close()
}
