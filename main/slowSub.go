package main

import (
	"fmt"
	"github.com/pebbe/zmq4"
	"log"
	"math/rand"
	"strconv"
	"time"
)

const MaxAllowedDelay = 1000 * time.Millisecond

func subscriberTask(pipe chan<- string) {
	subs, err := zmq4.NewSocket(zmq4.SUB)
	if err != nil {
		log.Fatalln(err)
	}
	subs.SetSubscribe("")
	subs.Connect("tcp://localhost:5556")
	defer subs.Close()

	for {
		msg, err := subs.RecvMessage(0)
		if err != nil {
			log.Fatalln(err)
		}

		i, err := strconv.Atoi(msg[0])
		if err != nil {
			log.Fatalln(err)
		}

		clock := time.Unix(int64(i), 0)
		fmt.Println(clock)

		if time.Now().After(clock.Add(MaxAllowedDelay)) {
			log.Println("Subscriber can't keep up,aborting")
			break
		}

		time.Sleep(time.Duration(1 + rand.Intn(2)))
	}

	pipe <- "gone and died"

}

func publisherTask(pipe <-chan string) {
	pub, err := zmq4.NewSocket(zmq4.PUB)
	if err != nil {
		log.Fatalln(err)
	}
	pub.Bind("tcp://*:5556")
	defer pub.Close()

LOOP:
	for {
		pub.SendMessage(time.Now().Unix())
		select {
		case <-pipe:
			break LOOP
		default:

		}
		time.Sleep(time.Millisecond)
	}
}

func SlowSubscriberDetection() {
	pubpipe := make(chan string)
	subpipe := make(chan string)

	go publisherTask(pubpipe)
	go subscriberTask(subpipe)
	<-subpipe

	pubpipe <- "break"
	time.Sleep(100 * time.Millisecond)
}
