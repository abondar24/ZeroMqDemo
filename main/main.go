package main

import (
	"github.com/abondar24/ZeroMqDemo/clients"
	"github.com/abondar24/ZeroMqDemo/servers"
	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/abondar24/ZeroMqDemo/brokers"
	"github.com/abondar24/ZeroMqDemo/queues"
	"github.com/abondar24/ZeroMqDemo/workers"
	"log"
	"os"
	"strconv"
)

var (
	base        = kingpin.New("base", "ZeroMQ Demo")
	hwSrv       = base.Command("hwserver", "Run HwServer")
	hwClt       = base.Command("hwclient", "Run HwClient")
	ver         = base.Command("version", "Show ZeroMQ version")
	wuSrv       = base.Command("wuserver", "Weather Update server")
	wuClt       = base.Command("wuclient", "Weather Update Client")
	wuCltArgs   = wuClt.Arg("zipcode", "Location").String()
	taskVent    = base.Command("taskVentilator", "Task Ventilator")
	taskWork    = base.Command("taskWorker", "Task Worker")
	taskSink    = base.Command("taskSink", "Task Sink")
	msReader    = base.Command("msreader", "Multi Socket Reader")
	msPoller    = base.Command("mspoller", "Another Multi Socket Reader")
	rrClient    = base.Command("rrclient", "Request Reply Client")
	rrBroker    = base.Command("rrbroker", "Request Reply Broker")
	rrWorker    = base.Command("rrworker", "Request Reply Worker")
	msqQueue    = base.Command("msgqueue", "Message Queue")
	interrupt   = base.Command("interrupt", "Interrupt Client")
	mtserver    = base.Command("mtserver", "Multithreaded hw server")
	mtrelay     = base.Command("mtrelay", "Multithreaded relay")
	syncpub     = base.Command("syncpub", "Synchronized publisher")
	syncsub     = base.Command("syncsub", "Synchronized subscriber")
	envpub      = base.Command("envpub", "Envelope publisher")
	envsub      = base.Command("envsub", "Envelope subscriber")
	id          = base.Command("identity", "Identities of sockets")
	rtreq       = base.Command("rtreq", "Router to request")
	llbroker    = base.Command("llbroker", "Load-Balancing broker")
	llbrokerr   = base.Command("llbrokerr", "Load-Balancing broker with reactor")
	asyncsrv    = base.Command("asyncsrv", "Async server")
	peering     = base.Command("peering", "Peering Broker")
	broker      = peering.Arg("broker", "Broker Name").Required().String()
	peers       = peering.Arg("peers", "List of peers").Required().Strings()
	lserver     = base.Command("ls", "Lazy server")
	lclient     = base.Command("lc", "Lazy client")
	rworker     = base.Command("rw", "Reliable worker")
	rqueue      = base.Command("rq", "Reliable queue")
	rrqueue     = base.Command("rrq", "Robust Reliable Queue")
	rorworker   = base.Command("rrw", "Robust reliable Worker")
	mdclient    = base.Command("mdcl", "Majordomo client")
	mdClVerbose = mdclient.Arg("verbose", "Verbose").String()
	mdworker    = base.Command("mdwr", "Majordomo worker")
	mdWrVerbose = mdworker.Arg("verbose", "Verbose").String()
	mdbroker    = base.Command("mdbr", "Majordomo worker")
	mdbrVerbose = mdbroker.Arg("verbose", "Verbose").String()
)

func main() {
	cmd := kingpin.MustParse(base.Parse(os.Args[1:]))

	switch cmd {

	case hwSrv.FullCommand():
		servers.HwServer()

	case hwClt.FullCommand():
		clients.HwClient()

	case ver.FullCommand():
		Version()

	case wuSrv.FullCommand():
		servers.WuServer()

	case wuClt.FullCommand():
		clients.WuClient(*wuCltArgs)

	case taskVent.FullCommand():
		servers.TaskVentilator()

	case taskWork.FullCommand():
		workers.TaskWorker()

	case taskSink.FullCommand():
		clients.TaskSink()

	case msReader.FullCommand():
		clients.MsReader()

	case msPoller.FullCommand():
		clients.MsPoller()

	case rrClient.FullCommand():
		clients.RRclient()

	case rrBroker.FullCommand():
		brokers.RRbroker()

	case rrWorker.FullCommand():
		workers.RequestReplyRworker()

	case msqQueue.FullCommand():
		queues.MsgQueue()

	case interrupt.FullCommand():
		clients.Interrupt()

	case mtserver.FullCommand():
		servers.MTserver()

	case mtrelay.FullCommand():
		MTrelay()

	case syncpub.FullCommand():
		servers.SyncPub()

	case syncsub.FullCommand():
		clients.SyncSub()

	case envpub.FullCommand():
		servers.EnvPub()

	case envsub.FullCommand():
		clients.EnvSub()

	case id.FullCommand():
		Identity()

	case rtreq.FullCommand():
		ReqRouter()

	case llbroker.FullCommand():
		brokers.LoadBalacningBroker()

	case llbrokerr.FullCommand():
		brokers.LoadBalacningReactorBroker()

	case asyncsrv.FullCommand():
		AsyncServer()

	case peering.FullCommand():
		Peering(*broker, *peers...)

	case lserver.FullCommand():
		servers.LazyServer()

	case lclient.FullCommand():
		clients.LazyClient()

	case rworker.FullCommand():
		workers.ReliableWorker()

	case rqueue.FullCommand():
		queues.ReliableQueue()

	case rrqueue.FullCommand():
		queues.RobustReliableQueue()

	case rorworker.FullCommand():
		workers.RobustReliableWorker()

	case mdclient.FullCommand():
		verbose, err := strconv.ParseBool(*mdClVerbose)
		if err == nil {
			log.Println(err)
		}

		clients.MajordomoClient(verbose)

	case mdworker.FullCommand():
		verbose, err := strconv.ParseBool(*mdWrVerbose)
		if err == nil {
			log.Println(err)
		}

		workers.MajordomoWorker(verbose)

	case mdbroker.FullCommand():
		verbose, err := strconv.ParseBool(*mdbrVerbose)
		if err == nil {
			log.Println(err)
		}

		brokers.MajordomoBroker(verbose)

	}

	os.Exit(0)
}
