package workers

import (
	"github.com/abondar24/ZeroMqDemo/mdapi"
	"log"
)

func MajordomoWorker(verbose bool) {

	session, err := mdapi.NewWorker("tcp://localhost:5555", "echo", verbose)
	if err != nil {
		log.Fatalln(err)
	}

	var reply []string
	for {
		request, err := session.SendReply(reply)
		if err != nil {
			log.Fatalln(err)
		}

		reply = request
	}

	log.Println(err)
}
