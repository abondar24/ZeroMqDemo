package main

import (
	"fmt"
	"github.com/pebbe/zmq4"
	"log"
	"os"
	"strconv"
)

const (
	ChunkSize = 250000
	PipeLine  = 10
)

func fileClient(pipe chan<- string) {
	dealer, err := zmq4.NewSocket(zmq4.DEALER)
	if err != nil {
		log.Fatalln(err)
	}
	dealer.Connect("tcp://localhost:6000")

	//max um chunks in transfer
	credit := PipeLine

	totalBytes := 0
	chunks := 0
	offset := 0

	for {

		for credit > 0 {
			dealer.SendMessage("fetch", totalBytes, ChunkSize)
			offset += ChunkSize
			credit--
		}

		chunk, err := dealer.RecvBytes(0)
		if err != nil {
			log.Fatalln(err)
		}

		chunks++
		credit++
		size := len(chunk)
		totalBytes += size
		if size < ChunkSize {
			break
		}
	}

	fmt.Printf("%v chunks received, %v bytes\n", chunks, totalBytes)
	pipe <- "OK"
}

func fileServer(filePath string) {
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalln(err)
	}

	router, err := zmq4.NewSocket(zmq4.ROUTER)
	if err != nil {
		log.Fatalln(err)
	}

	router.SetRcvhwm(PipeLine * 2)
	router.SetSndhwm(PipeLine * 2)
	router.Bind("tcp://*:6000")

	for {
		msg, err := router.RecvMessage(0)
		if err != nil {
			log.Fatalln(err)
		}

		id := msg[0]

		if msg[1] != "fetch" {
			log.Fatalln("command !='fetch'")
		}

		offset, err := strconv.ParseInt(msg[2], 10, 64)
		if err != nil {
			log.Fatalln(err)
		}

		maxChunkSize, err := strconv.Atoi(msg[3])
		if err != nil {
			log.Fatalln(err)
		}

		chunk := make([]byte, maxChunkSize)
		n, _ := file.ReadAt(chunk, offset)

		router.SendMessage(id, chunk[:n])

	}

	file.Close()
}

func FileTransfer(filePath string) {
	pipe := make(chan string)

	go fileServer(filePath)
	go fileClient(pipe)

	<-pipe
}
