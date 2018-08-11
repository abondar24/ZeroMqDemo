package kvmessage

import (
	"errors"
	"fmt"
	"github.com/pebbe/zmq4"
	"log"
	"os"
)

const (
	frameKey    = 0
	frameSeq    = 1
	frameBody   = 2
	kvmsgFrames = 3
)

type KVmsg struct {
	present []bool //presense inidcator for frame
	frame   []string
}

func NewKVMessage(seq int64) (kvmsg *KVmsg) {
	kvmsg = &KVmsg{
		present: make([]bool, kvmsgFrames),
		frame:   make([]string, kvmsgFrames),
	}

	kvmsg.SetSequence(seq)
	return
}

func RecvKVmsg(socket *zmq4.Socket) (kvmsg *KVmsg, err error) {
	kvmsg = &KVmsg{
		present: make([]bool, kvmsgFrames),
		frame:   make([]string, kvmsgFrames),
	}

	msg, err := socket.RecvMessage(0)
	if err != nil {
		return
	}

	for i := 0; i < kvmsgFrames && i < len(msg); i++ {
		kvmsg.frame[i] = msg[i]
		kvmsg.present[i] = true
	}

	return
}

func (kvmsg *KVmsg) SendKVmsg(socket *zmq4.Socket) (err error) {
	fmt.Printf("Send to %s: %q\n", socket, kvmsg.frame)
	_, err = socket.SendMessage(kvmsg.frame)
	return
}

func (kvmsg *KVmsg) GetKey() (key string, err error) {
	if !kvmsg.present[frameKey] {
		err = errors.New("key not set")
		return
	}

	key = kvmsg.frame[frameKey]
	return
}

func (kvmsg *KVmsg) SetKey(key string) {
	kvmsg.frame[frameKey] = key
	kvmsg.present[frameKey] = true
}

func (kvmsg *KVmsg) GetSequence() (sequence int64, err error) {
	if !kvmsg.present[frameSeq] {
		err = errors.New("sequence not set")
		return
	}

	source := kvmsg.frame[frameSeq]
	sequence = int64(source[0])<<56 +
		int64(source[1])<<48 +
		int64(source[2])<<40 +
		int64(source[3])<<32 +
		int64(source[4])<<24 +
		int64(source[5])<<16 +
		int64(source[6])<<8 +
		int64(source[7])

	return
}

func (kvmsg *KVmsg) SetSequence(sequence int64) {
	source := make([]byte, 8)
	source[0] = byte((sequence >> 56) & 255)
	source[1] = byte((sequence >> 48) & 255)
	source[2] = byte((sequence >> 40) & 255)
	source[3] = byte((sequence >> 32) & 255)
	source[4] = byte((sequence >> 24) & 255)
	source[5] = byte((sequence >> 18) & 255)
	source[6] = byte((sequence >> 8) & 255)
	source[7] = byte((sequence) & 255)

	kvmsg.frame[frameSeq] = string(source)
	kvmsg.present[frameSeq] = true
}

func (kvmsg *KVmsg) GetBody() (body string, err error) {
	if !kvmsg.present[frameBody] {
		err = errors.New("body not set")
		return
	}
	body = kvmsg.frame[frameBody]
	return
}

func (kvmsg *KVmsg) SetBody(body string) {
	kvmsg.frame[frameBody] = body
	kvmsg.present[frameBody] = true
}

// body size of last-read msg
func (kvmsg *KVmsg) BodySize() int {
	if kvmsg.present[frameBody] {
		return len(kvmsg.frame[frameBody])
	}

	return 0
}

func (kvmsg *KVmsg) StoreMsg(kvmap map[string]*KVmsg) {
	if kvmsg.present[frameKey] {
		kvmap[kvmsg.frame[frameKey]] = kvmsg
	}
}

func (kvmsg *KVmsg) Dump() {
	size := kvmsg.BodySize()
	body, err := kvmsg.GetBody()
	if err != nil {
		log.Fatalln(err)
	}

	seq, err := kvmsg.GetSequence()
	if err != nil {
		log.Fatalln(err)
	}

	key, err := kvmsg.GetKey()
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Fprintf(os.Stderr, "[seq:%v][key:%v][size:%v]", seq, key, size)
	for charNum := 0; charNum < size; charNum++ {
		fmt.Fprintf(os.Stderr, "%02X", body[charNum])
	}
	fmt.Fprintln(os.Stderr)
}
