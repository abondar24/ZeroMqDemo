package servers

import (
	"fmt"
	"github.com/abondar24/ZeroMqDemo/rpsapi"
	"github.com/pebbe/zmq4"
	"log"
	"strconv"
	"strings"
	"time"
)

type RelPsServer struct {
	kvmap     map[string]*rpsapi.KVmsg
	port      int
	sequence  int64
	snapshot  *zmq4.Socket
	publisher *zmq4.Socket
	collector *zmq4.Socket
}

func ReliablePubSubServer() {

	srv := &RelPsServer{
		port:  5556,
		kvmap: make(map[string]*rpsapi.KVmsg),
	}

	var err error
	//state request
	srv.snapshot, err = zmq4.NewSocket(zmq4.ROUTER)
	if err != nil {
		log.Fatalln(err)
	}
	srv.snapshot.Bind(fmt.Sprint("tcp://*:", srv.port))

	//updates
	srv.publisher, err = zmq4.NewSocket(zmq4.PUB)
	if err != nil {
		log.Fatalln(err)
	}
	srv.publisher.Bind(fmt.Sprint("tcp://*:", srv.port+1))

	//state updates
	srv.collector, err = zmq4.NewSocket(zmq4.PULL)
	if err != nil {
		log.Fatalln(err)
	}
	srv.collector.Bind(fmt.Sprint("tcp://*:", srv.port+2))

	reactor := zmq4.NewReactor()
	reactor.AddSocket(srv.snapshot, zmq4.POLLIN, func(st zmq4.State) error { return snapshots(srv) })
	reactor.AddSocket(srv.collector, zmq4.POLLIN, func(st zmq4.State) error { return collectors(srv) })
	reactor.AddChannelTime(time.Tick(1000*time.Millisecond), 1, func(v interface{}) error { return flushTTL(srv) })

	fmt.Println(reactor.Run(100 * time.Millisecond))
}

func snapshots(srv *RelPsServer) (err error) {
	msg, err := srv.snapshot.RecvMessage(0)
	if err != nil {
		return
	}

	id := msg[0]
	request := msg[1]
	if request != "ICANHAZ?" {
		fmt.Println("Bad request")
		return
	}

	subtree := msg[2]

	for _, kvmsg := range srv.kvmap {
		if key, _ := kvmsg.GetKey(); strings.HasPrefix(key, subtree) {
			srv.snapshot.Send(id, zmq4.SNDMORE)
			kvmsg.SendKVmsg(srv.snapshot)
		}
	}

	fmt.Printf("Sending state snaphot=%d\n", srv.sequence)
	srv.snapshot.Send(id, zmq4.SNDMORE)
	kvmsg := rpsapi.NewKVMessage(srv.sequence)
	kvmsg.SetKey("KTHXBAI")
	kvmsg.SetBody("")
	kvmsg.SendKVmsg(srv.snapshot)

	return
}

func collectors(srv *RelPsServer) (err error) {
	kvmsg, err := rpsapi.RecvKVmsg(srv.collector)
	if err != nil {
		return
	}
	srv.sequence++
	kvmsg.SetSequence(srv.sequence)
	kvmsg.SendKVmsg(srv.publisher)

	if ttls, e := kvmsg.GetProp("ttl"); e == nil {
		ttl, e := strconv.ParseInt(ttls, 10, 64)
		if e != nil {
			err = e
			return
		}
		kvmsg.SetProp("ttl", fmt.Sprint(time.Now().Add(time.Duration(ttl)*time.Second).Unix()))
	}

	kvmsg.StoreMsg(srv.kvmap)
	fmt.Println("Publishing state", srv.sequence)

	return
}

func flushTTL(srv *RelPsServer) (err error) {
	for _, kvmsg := range srv.kvmap {
		if ttls, e := kvmsg.GetProp("ttl"); e == nil {
			ttl, e := strconv.ParseInt(ttls, 10, 64)
			if e != nil {
				err = e
				continue
			}
			if time.Now().After(time.Unix(ttl, 0)) {
				srv.sequence++
				kvmsg.SetSequence(srv.sequence)
				kvmsg.SetBody("")
				e = kvmsg.SendKVmsg(srv.publisher)
				if e != nil {
					err = e
				}
				kvmsg.StoreMsg(srv.kvmap)
				fmt.Println("Publishing delete =", srv.sequence)
			}
		}
	}
	return
}
