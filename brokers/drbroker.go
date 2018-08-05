package brokers

import (
	"github.com/abondar24/ZeroMqDemo/mdapi"
	"github.com/pborman/uuid"
	"log"
	"os"

	"fmt"
	"io/ioutil"
	"strings"
	"time"
)

const TitanDir = ".titan"

func requestFilename(uuid string) string {
	return TitanDir + "/" + uuid + "req"
}

func replyFilename(uuid string) string {
	return TitanDir + "/" + uuid + "rep"
}

func handleRequest(chRequest chan<- string, verbose bool) {
	worker, err := mdapi.NewWorker("tcp://localhost:5555", "titan.request", verbose)
	if err != nil {
		log.Fatalln(err)
	}

	var reply []string

	for {
		request, err := worker.SendReply(reply)
		if err != nil {
			log.Fatalln(err)
		}

		os.MkdirAll(TitanDir, 0700)

		reqUUID := uuid.New()
		file, err := os.Create(requestFilename(reqUUID))
		if err != nil {
			log.Fatalln(err)
		}

		fmt.Fprint(file, strings.Join(request, "\n"))
		file.Close()

		chRequest <- reqUUID
		reply = []string{"200", reqUUID}
	}

}

func handleReply(verbose bool) {
	worker, err := mdapi.NewWorker("tcp://localhost:5555", "titan.reply", verbose)
	if err != nil {
		log.Fatalln(err)
	}

	pending := []string{"300"}
	unknown := []string{"400"}
	var reply []string

	for {
		request, err := worker.SendReply(reply)
		if err != nil {
			log.Fatalln(err)
		}

		reqUUID := request[0]
		reqFilename := requestFilename(reqUUID)
		repFilename := replyFilename(reqUUID)

		data, err := ioutil.ReadFile(repFilename)
		if err != nil {
			_, err = os.Stat(reqFilename)
			if err != nil {
				reply = unknown
			} else {
				reply = pending
			}

		} else {
			reply = strings.Split("200\n"+string(data), "\n")
		}
	}
}

func handleClose(verbose bool) {
	worker, err := mdapi.NewWorker("tcp://localhost:5555", "titan.close", verbose)
	if err != nil {
		log.Fatalln(err)
	}

	ok := []string{"200"}
	var reply []string
	for {
		request, err := worker.SendReply(reply)
		if err != nil {
			log.Fatalln(err)
		}

		reqUUID := request[0]
		os.Remove(requestFilename(reqUUID))
		os.Remove(replyFilename(reqUUID))

		reply = ok
	}

}

func isSuccess(entryUUID string, verbose bool) bool {
	_, err := os.Stat(replyFilename(entryUUID))
	if err != nil {
		return true
	}

	data, err := ioutil.ReadFile(requestFilename(entryUUID))
	if err != nil {
		return true
	}

	request := strings.Split(string(data), "\n")
	serviceName := request[0]
	request = request[1:]

	client, err := mdapi.NewMdClient("tcp://localhost:5555", verbose)
	if err != nil {
		log.Fatalln(err)
	}

	client.SetTimeout(time.Second)

	err = client.Send("mmi.service", serviceName)
	if err != nil {
		return false
	}

	mmiReply, err := client.Recv()
	if err != nil || mmiReply[0] != "200" {
		return false
	}

	err = client.Send(serviceName, request...)
	if err != nil {
		return false
	}

	reply, err := client.Recv()
	if err != nil {
		return false
	}

	file, err := os.Create(replyFilename(entryUUID))
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Fprint(file, strings.Join(reply, "\n"))
	file.Close()

	return true
}

func DisconnectedReliableBroker(verbose bool) {
	chRequest := make(chan string)

	go handleRequest(chRequest, verbose)
	go handleReply(verbose)
	go handleClose(verbose)

	os.Mkdir(TitanDir, 0700)

	queue := make([]string, 0)
	files, err := ioutil.ReadDir(TitanDir)
	if err != nil {
		log.Fatalln(err)
	}

	for _, file := range files {
		name := file.Name()
		if strings.HasSuffix(name, "req") {
			fileUUID := name[:len(name)-3]
			_, err := os.Stat(replyFilename(fileUUID))
			if err != nil {
				queue = append(queue, fileUUID)
			}
		}
	}

	for {
		select {
		case <-time.After(time.Second):
		case reqUUID := <-chRequest:
			queue = append(queue, reqUUID)
		}

		//brute-force dispatcher
		queue2 := make([]string, 0, len(queue))
		for _, entry := range queue {
			if verbose {
				fmt.Println("Processing request", entry)
			}

			if !isSuccess(entry, verbose) {
				queue2 = append(queue2, entry)
			}
		}
		queue = queue2
	}
}
