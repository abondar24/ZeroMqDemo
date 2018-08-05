package clients

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/abondar24/ZeroMqDemo/mdapi"
	"time"
)

func serviceCall(session *mdapi.MdClient, service string, request ...string) (reply []string, err error) {

	reply = []string{}

	err = session.Send(service, request...)
	if err != nil {
		log.Fatalln(err)
	}

	msg, err := session.Recv()

	if err == nil {
		switch status := msg[0]; status {
		case "200":
			reply = msg[1:]
			return
		case "400":
			fmt.Println("Client fatal error")
			os.Exit(1)
		case "500":
			fmt.Println("Server fatal error")
			os.Exit(1)

		}
	} else {
		log.Fatalln(err)
	}

	err = errors.New("did not succeed")
	return
}

func DisconnectedReliableClient(verbose bool) {
	session, err := mdapi.NewMdClient("tcp://localhost:5555", verbose)
	if err != nil {
		log.Fatalln(err)
	}

	reply, err := serviceCall(session, "titan.request", "echo", "hi")
	if err != nil {
		log.Fatalln(err)
	}

	uuid := reply[0]
	fmt.Println("Request uuid", uuid)

	time.Sleep(100 * time.Millisecond)

	for {
		reply, err = serviceCall(session, "titan.reply", uuid)
		if err == nil {
			fmt.Println("Reply:", reply[0])

			serviceCall(session, "titan.close", uuid)
			break
		} else {
			fmt.Println("No reply yet")
			time.Sleep(5 * time.Second)
		}
	}
}
