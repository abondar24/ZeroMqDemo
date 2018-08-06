package servers

import (
	"fmt"
	"github.com/abondar24/ZeroMqDemo/haServerApi"
	"github.com/pebbe/zmq4"
	"log"
)

func echo(socket *zmq4.Socket) (err error) {

	msg, err := socket.RecvMessage(0)
	if err != nil {
		return
	}

	_, err = socket.SendMessage(msg)

	return
}

func HAServer(isPrimary bool) {

	var hsrv *haServerApi.HAServer
	var err error
	if isPrimary {
		fmt.Println("Primary active, waiting for backup (passive)")
		hsrv, err = haServerApi.NewServer(haServerApi.Primary, "tcp://*:5003", "tcp://localhost:5004")
		if err != nil {
			log.Fatalln(err)
		}

		hsrv.Voter("tcp://*:5001", zmq4.ROUTER, echo)

	} else if !isPrimary {
		fmt.Println("Backup passive, waiting for primary (active)")
		hsrv, err = haServerApi.NewServer(haServerApi.Backup, "tcp://*:5004", "tcp://localhost:5003")
		if err != nil {
			log.Fatalln(err)
		}

		hsrv.Voter("tcp://*:5002", zmq4.ROUTER, echo)
	}

	hsrv.StartReactor()
}
