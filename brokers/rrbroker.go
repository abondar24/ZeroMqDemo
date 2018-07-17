package brokers

import (
	"github.com/pebbe/zmq4"
	"log"
)

func RRbroker(){
	frontend,err := zmq4.NewSocket(zmq4.ROUTER)
	if err!=nil{
		log.Fatal(err)
	}
	defer frontend.Close()
	frontend.Bind("tcp://*:5559")

	backend,err := zmq4.NewSocket(zmq4.DEALER)
	if err!=nil{
		log.Fatal(err)
	}
	defer backend.Close()
	backend.Bind("tcp://*:5560")

	poller := zmq4.NewPoller()
	poller.Add(frontend,zmq4.POLLIN)
	poller.Add(backend,zmq4.POLLIN)

	for {
		sockets,err := poller.Poll(-1)
		if err!=nil{
			log.Println(err)
		}

		for _,socket :=range sockets{
			switch s:=socket.Socket;s {
			case frontend:
				for {
					msg,err := s.Recv(0)
					if err!=nil{
						log.Println(err)
					}

					if more,_ := s.GetRcvmore();more{
						backend.Send(msg,zmq4.SNDMORE)
					} else {
						backend.Send(msg,0)
						break
					}
				}
			case backend:
				for {
					msg,err := s.Recv(0)
					if err!=nil{
						log.Println(err)
					}

					if more,_ := s.GetRcvmore();more{
						frontend.Send(msg,zmq4.SNDMORE)
					} else {
						frontend.Send(msg,0)
						break
					}
				}
			}
		}

	}
}
