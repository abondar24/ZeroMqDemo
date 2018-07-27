package workers

import (
	"fmt"
	"github.com/pebbe/zmq4"
	"log"
	"math/rand"
	"time"
)

const WorkerReady = "\001"

func ReliableWorker() {
	worker, err := zmq4.NewSocket(zmq4.REQ)
	if err != nil {
		log.Fatalln(err)
	}
	defer worker.Close()

	rand.Seed(time.Now().UnixNano())
	id := fmt.Sprintf("%04X-%04X", rand.Intn(0x10000), rand.Intn(0x10000))
	worker.SetIdentity(id)
	worker.Connect("tcp://localhost:5556")

	fmt.Printf("(%s) worker ready\n", id)
	worker.Send(WorkerReady, 0)

	for cycles := 0; true; {
		msg, err := worker.RecvMessage(0)
		if err != nil {
			log.Fatalln(err)
		}
		cycles++

		//simulate issues
		if cycles > 4 && rand.Intn(5) == 0 {
			fmt.Println("Simulating crash")
			break
		} else if cycles > 4 && rand.Intn(5) == 0 {
			fmt.Println("(%s) Simulate CPU overload", id)
			time.Sleep(3 * time.Second)
		}

		fmt.Printf("Normal reply (%s)\n", id)
		time.Sleep(time.Second)
		worker.SendMessage(msg)

	}
}
