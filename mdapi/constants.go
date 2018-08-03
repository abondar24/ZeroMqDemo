package mdapi

import "time"

const (
	MdClientVer = "MD_CL_01"
	MdWorkerVer = "MD_WR_01"

	MdReady = string(iota + 1)
	MdRequest
	MdReply
	MdHearbeart
	MdDisconnect
	HeartbeatInterval = 2500 * time.Millisecond
	HeartbeatExpiry   = HeartbeatInterval * HeartbeatLiveness
)

var (
	MdCommands = map[string]string{
		MdReady:      "READY",
		MdRequest:    "REQUEST",
		MdReply:      "REPLY",
		MdHearbeart:  "HEARTBEAT",
		MdDisconnect: "DISCONNECT",
	}
)
