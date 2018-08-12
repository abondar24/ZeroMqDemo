package servers

import (
	"fmt"
	"github.com/abondar24/ZeroMqDemo/haServerApi"
	"github.com/abondar24/ZeroMqDemo/rpsapi"
	"github.com/pebbe/zmq4"
	"log"
	"strconv"
	"strings"
	"time"
)

type RelPsServer struct {
	kvmap      map[string]*rpsapi.KVmsg
	kvmapInit  bool
	hserver    *haServerApi.HAServer
	port       int
	sequence   int64
	peer       int
	subscriber *zmq4.Socket
	publisher  *zmq4.Socket
	collector  *zmq4.Socket
	pending    []*rpsapi.KVmsg
	primary    bool
	active     bool
	passive    bool
}

func ReliablePubSubServer(isPrimary bool) {

	srv := &RelPsServer{}

	var err error

	if isPrimary {
		fmt.Println("Primary active, waiting for backup (passive)")
		srv.hserver, err = haServerApi.NewServer(haServerApi.Primary, "tcp://*:5003", "tcp://localhost:5004")
		if err != nil {
			log.Fatalln(err)
		}

		srv.hserver.Voter("tcp://*:5556", zmq4.ROUTER, func(soc *zmq4.Socket) error { return snapshots(soc, srv) })
		srv.port = 5556
		srv.peer = 5566
		srv.primary = true

	} else {
		fmt.Println("Backup passive, waiting for primary (active)")
		srv.hserver, err = haServerApi.NewServer(haServerApi.Backup, "tcp://*:5004", "tcp://localhost:5003")
		if err != nil {
			log.Fatalln(err)
		}

		srv.hserver.Voter("tcp://*:5566", zmq4.ROUTER, func(soc *zmq4.Socket) error { return snapshots(soc, srv) })
		srv.port = 5566
		srv.peer = 5556
		srv.primary = false
	}

	if srv.primary {
		srv.kvmap = make(map[string]*rpsapi.KVmsg, 0)
		srv.kvmapInit = true
	}

	srv.pending = make([]*rpsapi.KVmsg, 0)
	srv.hserver.SetVerbose(true)

	srv.publisher, err = zmq4.NewSocket(zmq4.PUB)
	if err != nil {
		log.Fatalln(err)
	}
	srv.publisher.Bind(fmt.Sprint("tcp://*:", srv.port+1))

	//state updates
	srv.collector, err = zmq4.NewSocket(zmq4.SUB)
	if err != nil {
		log.Fatalln(err)
	}
	srv.collector.SetSubscribe("")
	srv.collector.Bind(fmt.Sprint("tcp://*:", srv.port+2))

	//client interface for peering
	srv.subscriber, err = zmq4.NewSocket(zmq4.SUB)
	if err != nil {
		log.Fatalln(err)
	}
	srv.subscriber.SetSubscribe("")
	srv.subscriber.Connect(fmt.Sprint("tcp://localhost:", srv.peer+1))

	srv.hserver.NewActive(func() error { return newActive(srv) })
	srv.hserver.NewPassive(func() error { return newPassive(srv) })

	srv.hserver.Reactor.AddSocket(srv.collector, zmq4.POLLIN, func(st zmq4.State) error { return collectors(srv) })
	srv.hserver.Reactor.AddChannelTime(time.Tick(1000*time.Millisecond), 1, func(v interface{}) error {

		if e := flushTTL(srv); e != nil {
			return e
		}
		return sendHugz(srv)
	})

	err = srv.hserver.StartReactor()
	if err != nil {
		log.Println(err)
	}

}

func snapshots(socket *zmq4.Socket, srv *RelPsServer) (err error) {
	msg, err := socket.RecvMessage(0)
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
			socket.Send(id, zmq4.SNDMORE)
			kvmsg.SendKVmsg(socket)
		}
	}

	fmt.Printf("Sending state snaphot=%d\n", srv.sequence)
	socket.Send(id, zmq4.SNDMORE)
	kvmsg := rpsapi.NewKVMessage(srv.sequence)
	kvmsg.SetKey("KTHXBAI")
	kvmsg.SetBody(subtree)
	kvmsg.SendKVmsg(socket)

	return
}

func collectors(srv *RelPsServer) (err error) {
	kvmsg, err := rpsapi.RecvKVmsg(srv.collector)
	if err != nil {
		return
	}

	if srv.active {
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

	} else {
		//  If message already from active, drop it, else hold on pending list
		if !srv.wasPending(kvmsg) {
			srv.pending = append(srv.pending, kvmsg)
		}
	}

	return
}

//  If message was already on pending list, remove it and return TRUE, else return FALSE.
func (srv *RelPsServer) wasPending(kvmsg *rpsapi.KVmsg) bool {
	uuid1, _ := kvmsg.GetUUID()

	for i, msg := range srv.pending {
		if uuid2, _ := msg.GetUUID(); uuid1 == uuid2 {
			srv.pending = append(srv.pending[:i], srv.pending[i+1:]...)
			return true
		}
	}

	return false
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

//send HUGZ for server state detection
func sendHugz(srv *RelPsServer) (err error) {
	kvmsg := rpsapi.NewKVMessage(srv.sequence)
	kvmsg.SetKey("HUGZ")
	kvmsg.SetBody("")
	err = kvmsg.SendKVmsg(srv.publisher)
	return
}

func newActive(srv *RelPsServer) (err error) {
	srv.active = true
	srv.passive = false

	srv.hserver.Reactor.RemoveSocket(srv.subscriber)

	for _, msg := range srv.pending {
		srv.sequence++
		msg.SetSequence(srv.sequence)
		msg.SendKVmsg(srv.publisher)
		msg.StoreMsg(srv.kvmap)
		fmt.Println("Publishing pending =", srv.sequence)
	}

	srv.pending = srv.pending[0:0]
	return
}

func newPassive(srv *RelPsServer) (err error) {
	srv.kvmap = make(map[string]*rpsapi.KVmsg)
	srv.kvmapInit = false
	srv.active = false
	srv.passive = true

	srv.hserver.Reactor.AddSocket(srv.subscriber, zmq4.POLLIN, func(st zmq4.State) error { return subscriber(srv) })

	return
}

//state update
func subscriber(srv *RelPsServer) (err error) {
	if !srv.kvmapInit {
		srv.kvmapInit = true
		snapshot, err := zmq4.NewSocket(zmq4.DEALER)
		if err != nil {
			log.Fatalln(err)
		}

		snapshot.Connect(fmt.Sprint("tcp://localhost:", srv.peer))
		fmt.Printf("Asking for snaphot from: tcp://localhost:%v\n", srv.peer)
		snapshot.SendMessage("ICANHAZ?", "")
		for {
			kvmsg, e := rpsapi.RecvKVmsg(snapshot)
			if e != nil {
				err = e
				break
			}

			if key, _ := kvmsg.GetKey(); key == "KTHXBAI" {
				srv.sequence, _ = kvmsg.GetSequence()
				break
			}

			kvmsg.StoreMsg(srv.kvmap)
		}
		fmt.Println("Received snapshot =", srv.sequence)
	}

	kvmsg, e := rpsapi.RecvKVmsg(srv.subscriber)
	if e != nil {
		err = e
		return
	}

	if key, _ := kvmsg.GetKey(); key != "HUGZ" {
		if !srv.wasPending(kvmsg) {
			//  If active update came before client update, flip it
			//  around, store active update (with sequence) on pending
			//  list and use to clear client update when it comes later
			srv.pending = append(srv.pending, kvmsg)
		}

		if seq, _ := kvmsg.GetSequence(); seq > srv.sequence {
			srv.sequence = seq
			kvmsg.StoreMsg(srv.kvmap)
			fmt.Println("Received update =", srv.sequence)
		}
	}

	return
}
