package libbitcoin

import (
	zmq "github.com/pebbe/zmq4"
	"encoding/binary"
	"math/rand"
)
const MAX_UNIT32 = 4294967295

type ClientBase struct {
	socket         *ZMQSocket
	subscriptions  map[int] func(interface{})
	messages       [][]byte
	handler        chan Response

}

func NewClientBase(address string, publicKey string) *ClientBase {
	handler := make(chan Response)
	subcriptions := make(map[int] func(interface{}))
	cb := ClientBase{
		socket: NewSocket(handler, zmq.DEALER),
		handler: handler,
		subscriptions: subcriptions,
		messages: [][]byte{},
	}
	cb.socket.Connect(address, publicKey)
	go cb.handleResponse()
	return &cb
}

func (cb *ClientBase) SendCommand(command string, data []byte, callback func(interface{})) {
	txid := rand.Intn(MAX_UNIT32)
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, uint32(txid))

	cb.subscriptions[txid] = callback

	cb.socket.Send([]byte(command), 2)
	cb.socket.Send(b, 2)
	cb.socket.Send(data, 0)
}

func (cb *ClientBase) messageReceived(command string, id, data []byte){
	txid := int(binary.LittleEndian.Uint32(id))
	ParseResponse(command, data, cb.subscriptions[txid])
}

func (cb *ClientBase) handleResponse(){
	for r := range cb.handler {
		cb.messages = append(cb.messages, r.data)
		if !r.more {
			command := string(cb.messages[0])
			id := cb.messages[1]
			data := cb.messages[2]
			cb.messageReceived(command, id, data)
			cb.messages = [][]byte{}
		}
	}
}


