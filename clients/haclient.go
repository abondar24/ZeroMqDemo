package clients

import (
	"fmt"
	"github.com/pebbe/zmq4"
	"log"
	"strconv"
	"time"
)

const (
	ReqTimeout  = 1000 * time.Millisecond
	SettleDelay = 2000 * time.Millisecond
)

func HighlyAvailableClient() {
	server := []string{"tcp://localhost:5001", "tcp://localhost:5002"}

	serverNum := 0

	fmt.Printf("Connecting to server at %s...", server[serverNum])

	client, err := zmq4.NewSocket(zmq4.REQ)
	if err != nil {
		log.Fatalln(err)
	}

	client.Connect(server[serverNum])

	poller := zmq4.NewPoller()
	poller.Add(client, zmq4.POLLIN)

	sequence := 0

LOOP:
	for {
		sequence++
		client.SendMessage(sequence)

		for expectReply := true; expectReply; {
			polled, err := poller.PollAll(ReqTimeout)
			if err != nil {
				log.Println(err)
				break LOOP
			}

			if len(polled) == 1 {
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
					expectReply = false
					time.Sleep(time.Second)
				} else {
					fmt.Printf("Bad reply from server: %q\n", reply)
				}
			} else {
				fmt.Println("No response from the server. Failing over")

				client.Close()
				serverNum = 1 - serverNum
				time.Sleep(SettleDelay)
				fmt.Printf("Connecting to server at %s...", server[serverNum])

				client, err := zmq4.NewSocket(zmq4.REQ)
				if err != nil {
					log.Fatalln(err)
				}

				client.Connect(server[serverNum])
				poller := zmq4.NewPoller()
				poller.Add(client, zmq4.POLLIN)
				client.SendMessage(sequence)
			}

		}
	}
}
