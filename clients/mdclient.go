package clients

import (
	"fmt"
	"github.com/abondar24/ZeroMqDemo/mdapi"
	"log"
)

func MajordomoClient(verbose bool) {

	session, err := mdapi.NewMdClient("tcp://localhost:5555", verbose)
	if err != nil {
		log.Fatalln(err)
	}

	var count int
	for count = 0; count < 100000; count++ {
		err := session.Send("echo", "Hello World")
		if err != nil {
			log.Fatalln(err)
		}
	}

	for count = 0; count < 100000; count++ {
		_, err := session.Recv()
		if err != nil {
			log.Fatalln(err)
		}
	}

	fmt.Printf("%d replies received\n", count)
}
