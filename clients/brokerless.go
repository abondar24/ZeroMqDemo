package clients

import (
	"errors"
	"fmt"
	flcapi "github.com/abondar24/ZeroMqDemo/failapi"
	"github.com/pebbe/zmq4"
	"log"
	"strconv"
	"time"
)

const (
	Timeout       = 1000 * time.Millisecond
	MaxRetries    = 3
	GlobalTimeout = 2500 * time.Millisecond
)

func tryRequest(endpoint string, request []string) (reply []string, err error) {
	fmt.Printf("Trying echo service at %s...\n", endpoint)

	client, err := zmq4.NewSocket(zmq4.REQ)

	client.Connect(endpoint)
	client.SendMessage(request)

	poller := zmq4.NewPoller()
	poller.Add(client, zmq4.POLLIN)
	polled, err := poller.Poll(Timeout)

	reply = []string{}
	if len(polled) == 1 {
		reply, err = client.RecvMessage(0)
	} else {
		err = errors.New("time out")
	}

	return
}

func BrokerlessFailoverClient(endpoints ...string) {
	request := []string{"hiiii"}
	var reply []string

	var err error

	endpointsSize := len(endpoints) - 1

	if endpointsSize == 0 {
		fmt.Printf("Syntax: %s <endpoint> ...\n", endpoints[0])
	} else if endpointsSize == 1 {
		for retries := 0; retries < MaxRetries; retries++ {
			endpoint := endpoints[1]
			reply, err = tryRequest(endpoint, request)
			if err == nil {
				break
			}

			fmt.Printf("No response from %s, retrying...", endpoint)
		}
	} else {
		for endpointNum := 0; endpointNum < endpointsSize; endpointNum++ {
			endpoint := endpoints[endpointNum+1]
			reply, err = tryRequest(endpoint, request)
			if err == nil {
				break
			}
			fmt.Print("No response from", endpoint)
		}
	}

	if len(reply) > 0 {
		fmt.Printf("Service is running ok: %q\n", reply)
	}
}

func BrokerlessFailoverReplyClient(endpoints ...string) {
	replyClient := newFlRepClient()

	for endpoint := 1; endpoint < len(endpoints); endpoint++ {
		replyClient.connect(endpoints[endpoint])
	}

	start := time.Now()
	for requests := 10000; requests > 0; requests-- {
		_, err := replyClient.request("random name")
		if err != nil {
			log.Println("name service not available, aborting")
			log.Fatalln(err)
		}
	}

	fmt.Println("Avg round trip costL", time.Now().Sub(start))
}

type FaiReplyClient struct {
	socket   *zmq4.Socket
	servers  int
	sequence int
}

func newFlRepClient() (clt *FaiReplyClient) {
	clt = &FaiReplyClient{}

	clt.socket, _ = zmq4.NewSocket(zmq4.DEALER)
	return
}

func (clt *FaiReplyClient) connect(endpoint string) {
	clt.socket.Connect(endpoint)
	clt.servers++
}

func (clt *FaiReplyClient) request(request ...string) (reply []string, err error) {
	reply = []string{}

	clt.sequence++

	for srv := 0; srv < clt.servers; srv++ {
		clt.socket.SendMessage("", clt.sequence, request)
	}

	//wait for a matching reply
	endTime := time.Now().Add(GlobalTimeout)
	poller := zmq4.NewPoller()
	poller.Add(clt.socket, zmq4.POLLIN)

	for time.Now().Before(endTime) {
		polled, err := poller.Poll(endTime.Sub(time.Now()))
		if err == nil && len(polled) > 0 {
			reply, _ = clt.socket.RecvMessage(0)
			if len(reply) != 3 {
				log.Fatalln("len(reply) != 3")
			}
			sequence := reply[1]
			reply = reply[2:]
			sequenceNum, _ := strconv.Atoi(sequence)
			if sequenceNum == clt.sequence {
				break
			}
		}
	}

	if len(reply) == 0 {
		err = errors.New("no reply")
	}

	return
}

func BrokerlessFailoverRoutingClient() {
	client := flcapi.New()

	client.Connect("tcp://localhost:5555")
	client.Connect("tcp://localhost:5556")
	client.Connect("tcp://localhost:5557")

	start := time.Now()
	req := []string{"random name"}

	for requests := 1000; requests > 0; requests-- {
		_, err := client.Request(req)
		if err != nil {
			fmt.Println("Name service is down")
			log.Fatalln(err)
		}

	}

	fmt.Println("Average round trip cost:", time.Now().Sub(start)/1000)
}
