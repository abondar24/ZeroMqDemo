package rpsapi

import (
	"github.com/pebbe/zmq4"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestKVmsg(t *testing.T) {
	output, err := zmq4.NewSocket(zmq4.DEALER)
	if err != nil {
		t.Error(err)
	}

	err = output.Bind("ipc://kvmsg_test.ipc")
	if err != nil {
		t.Error(err)
	}

	input, err := zmq4.NewSocket(zmq4.DEALER)
	if err != nil {
		t.Error(err)
	}

	err = input.Connect("ipc://kvmsg_test.ipc")
	if err != nil {
		t.Error(err)
	}

	kvmap := make(map[string]*KVmsg)

	kvmsg := NewKVMessage(1)
	kvmsg.SetKey("key")
	kvmsg.SetUUID()
	kvmsg.SetBody("body")
	kvmsg.SetProp("pr", "val")
	kvmsg.Dump()
	err = kvmsg.SendKVmsg(output)

	kvmsg.StoreMsg(kvmap)
	if err != nil {
		t.Error(err)
	}

	kvmsg, err = RecvKVmsg(input)
	if err != nil {
		t.Error(err)
	}

	kvmsg.Dump()
	key, err := kvmsg.GetKey()
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, "key", key)

	prop, err := kvmsg.GetProp("pr")
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, "val", prop)

	kvmsg.StoreMsg(kvmap)
	input.Close()
	output.Close()
	os.Remove("kvmsg_test.ipc")
}
