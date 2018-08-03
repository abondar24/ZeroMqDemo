package clients

import (
	"fmt"
	"github.com/abondar24/ZeroMqDemo/mdapi"
	"log"
)

func MajordomoDiscoverySearch(verbose bool) {
	session, err := mdapi.NewMdClient("tcp://localhost:5555", verbose)
	if err != nil {
		log.Fatalln(err)
	}

	request := "echo"

	err = session.Send("mmi.service", request)
	if err != nil {
		log.Fatalln(err)
	}

	reply, err := session.Recv()
	if err == nil {
		fmt.Println("Lookup echo service:", reply[0])
	} else {
		log.Fatalln(err)
	}
}
