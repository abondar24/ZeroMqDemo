package servers

import (
	"fmt"
	"github.com/pebbe/zmq4"
	"log"
)

func BrokerlessFailoverServer(endpoint string) {

	server, err := zmq4.NewSocket(zmq4.REP)
	if err != nil {
		log.Fatalln(err)
	}

	server.Bind(endpoint)

	fmt.Println("Echo service is ready at", endpoint)
	for {
		msg, err := server.RecvMessage(0)
		if err != nil {
			log.Fatalln(err)
		}

		server.SendMessage(msg)
	}

	fmt.Println("Interrupted")
}

func BrokerlessFailoverMsgReplyServer(endpoint string) {

	server, err := zmq4.NewSocket(zmq4.REP)
	if err != nil {
		log.Fatalln(err)
	}

	server.Bind(endpoint)

	fmt.Println("Echo service is ready at", endpoint)
	for {
		request, err := server.RecvMessage(0)
		if err != nil {
			log.Fatalln(err)
		}

		if len(request) != 2 {
			log.Fatalln("len(request) != 2")
		}

		id := request[0]

		server.SendMessage(id, "OK")
	}

	fmt.Println("Interrupted")
}

func BrokerlessRoutingServer(port string, verbose bool) {
	bindEndpoint := "tcp://*:" + port
	connectEndpoint := "tcp://localhost:" + port

	server, err := zmq4.NewSocket(zmq4.ROUTER)
	if err != nil {
		log.Fatalln(err)
	}

	server.SetIdentity(connectEndpoint)
	server.Bind(bindEndpoint)
	fmt.Println("Server ready at", bindEndpoint)

	for {
		request, err := server.RecvMessage(0)
		if err != nil {
			log.Fatalln(err)
		}

		if verbose {
			fmt.Printf("%q\n", request)
		}

		//  Frame 0: identity of client
		//  Frame 1: PING, or client control frame
		//  Frame 2: request body
		id := request[0]
		control := request[1]
		reply := make([]string, 1, 3)

		if control == "PING" {
			reply = append(reply, "PONG")
		} else {
			reply = append(reply, control)
			reply = append(reply, "OK")
		}

		reply[0] = id
		if verbose {
			fmt.Printf("%q\n", reply)
		}

		server.SendMessage(reply)
	}

	fmt.Println("Interrupted")
}
