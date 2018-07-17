package clients

import (
	"log"
	"fmt"
	"github.com/pebbe/zmq4"
	"strings"
	"strconv"
)

func WuClient(zipcode string){

	socket, err := zmq4.NewSocket(zmq4.SUB)
	if err!=nil{
		log.Fatalln(err)
	}

	defer  socket.Close()

	var temps []string
	var temp int64
	totalTemp := 0
	filter := "76137"

	if zipcode!="" {
		filter = zipcode
	}


	fmt.Printf("Collecting weather updates for %s...\n",filter)
	if socket.SetSubscribe(filter)!= nil {
		log.Fatalln(err)
	}

	if socket.Connect("tcp://localhost:5556")!= nil {
		log.Fatalln(err)
	}

	for i:=0; i<101; i++{
		datapt,err := socket.Recv(0)
		if err!=nil{
			log.Println(err)
		}

		temps = strings.Split(string(datapt)," ")
		temp,err = strconv.ParseInt(temps[1],10,64)
		if err != nil {
			log.Println(err)

		}
		totalTemp += int(temp)
	}

	fmt.Printf("Average temperature for location %s was %dC \n\n",filter,totalTemp/100)
}
