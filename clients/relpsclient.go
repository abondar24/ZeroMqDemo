package clients

import (
	"fmt"
	"github.com/abondar24/ZeroMqDemo/rpsapi"
	"log"
	"math/rand"
	"time"
)

const Subtree = "/client/"

func ReliablePubSubClient() {
	client := rpsapi.NewClient()

	client.Subtree(Subtree)
	client.Connect("tcp://localhost", "5556")
	client.Connect("tcp://localhost", "5566")

	for {
		key := fmt.Sprintf("%s%d", Subtree, rand.Intn(10000))
		value := fmt.Sprint(rand.Intn(1000000))
		client.Set(key, value, rand.Intn(30))

		v, err := client.Get(key)
		if err != nil {
			log.Fatalln(err)
		}

		if v != value {
			log.Fatalf("Set: %v - Get: %v - Equal: %v\n", value, v, value == v)
		}
		time.Sleep(time.Second)
	}
}
