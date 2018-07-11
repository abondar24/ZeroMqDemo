package main

import (
	"gopkg.in/alecthomas/kingpin.v2"
	"github.com/abondar24/ZeroMqDemo/clients"
	"github.com/abondar24/ZeroMqDemo/servers"

	"os"
	"github.com/abondar24/ZeroMqDemo/workers"
)

var (
	base  = kingpin.New("base", "ZeroMQ Demo")
	hwSrv = base.Command("hwserver", "Run HwServer")
	hwClt = base.Command("hwclient", "Run HwClient")
	ver   = base.Command("version", "Show ZeroMQ version")
	wuSrv = base.Command("wuserver","Weather Update server")
	wuClt = base.Command("wuclient","Weather Update Client")
	wuCltArgs = wuClt.Arg("zipcode","Location").String()
	taskVent = base.Command("taskVentilator","Task Ventilator")
	taskWork = base.Command("taskWorker","Task Worker")
	taskSink = base.Command("taskSink","Task Sink")

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
	}

	os.Exit(0)
}
