package main

import (
	"github.com/pebbe/zmq4"
	"fmt"
)

func Version(){
	major,minor,patch := zmq4.Version()
	fmt.Printf("Current 0MQ version is %d.%d.%d\n",major,minor,patch)
}
