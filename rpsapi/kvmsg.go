package rpsapi

import (
	"errors"
	"fmt"
	"github.com/pborman/uuid"
	"github.com/pebbe/zmq4"
	"log"
	"os"
	"strings"
)

const (
	frameKey    = 0
	frameSeq    = 1
	frameUUID   = 2
	frameProps  = 3
	frameBody   = 4
	kvmsgFrames = 5
)

type KVmsg struct {
	present []bool //presense inidcator for frame
	frame   []string
	props   []string
}

func NewKVMessage(seq int64) (kvmsg *KVmsg) {
	kvmsg = &KVmsg{
		present: make([]bool, kvmsgFrames),
		frame:   make([]string, kvmsgFrames),
		props:   make([]string, 0),
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
	kvmsg.decodeProps()

	return
}

func (kvmsg *KVmsg) encodeProps() {
	kvmsg.frame[frameProps] = strings.Join(kvmsg.props, "\n") + "\n"
	kvmsg.present[frameProps] = true
}

func (kvmsg *KVmsg) decodeProps() {
	kvmsg.props = strings.Split(kvmsg.frame[frameProps], "\n")
	if ln := len(kvmsg.props); ln > 0 && kvmsg.props[ln-1] == "" {
		kvmsg.props = kvmsg.props[:ln-1]
	}
}

func (kvmsg *KVmsg) SendKVmsg(socket *zmq4.Socket) (err error) {
	fmt.Printf("Send to %s: %q\n", socket, kvmsg.frame)
	kvmsg.encodeProps()
	_, err = socket.SendMessage(kvmsg.frame)
	return
}

func (kvmsg *KVmsg) Dup() (dup *KVmsg) {
	dup = &KVmsg{
		present: make([]bool, kvmsgFrames),
		frame:   make([]string, kvmsgFrames),
		props:   make([]string, len(kvmsg.props)),
	}
	copy(dup.present, kvmsg.present)
	copy(dup.frame, kvmsg.frame)
	copy(dup.props, kvmsg.props)
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

func (kvmsg *KVmsg) GetUUID() (uuid string, err error) {
	if !kvmsg.present[frameUUID] {
		err = errors.New("uuid not set")
		return
	}

	uuid = kvmsg.frame[frameUUID]
	return
}

func (kvmsg *KVmsg) SetUUID() {
	kvmsg.frame[frameUUID] = string(uuid.NewRandom())
	kvmsg.present[frameUUID] = true
}

func (kvmsg *KVmsg) GetProp(name string) (value string, err error) {
	if !kvmsg.present[frameProps] {
		err = errors.New("no properties set")
		return
	}

	f := name + "="
	for _, prop := range kvmsg.props {
		if strings.HasPrefix(prop, f) {
			value = prop[len(f):]
			return
		}
	}
	err = errors.New("property not set")
	return
}

func (kvmsg *KVmsg) SetProp(name, value string) (err error) {
	if strings.Index(name, "=") >= 0 {
		err = errors.New("no '=' allowed in property name")
		return
	}

	p := name + "="
	for i, prop := range kvmsg.props {
		if strings.HasPrefix(prop, p) {
			kvmsg.props = append(kvmsg.props[:i], kvmsg.props[i+1:]...)
			break
		}
	}

	kvmsg.props = append(kvmsg.props, name+"="+value)
	kvmsg.present[frameProps] = true
	return
}

func (kvmsg *KVmsg) StoreMsg(kvmap map[string]*KVmsg) {
	if kvmsg.present[frameKey] {
		if kvmsg.present[frameBody] && kvmsg.frame[frameBody] != "" {
			kvmap[kvmsg.frame[frameKey]] = kvmsg
		} else {
			delete(kvmap, kvmsg.frame[frameKey])
		}

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
	p := "["
	for _, prop := range kvmsg.props {
		fmt.Fprint(os.Stderr, p, prop)
		p = ";"
	}
	if p == ";" {
		fmt.Fprint(os.Stderr, "]")
	}

	for charNum := 0; charNum < size; charNum++ {
		fmt.Fprintf(os.Stderr, "%02X", body[charNum])
	}
	fmt.Fprintln(os.Stderr)
}
