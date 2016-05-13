package libbitcoin

import (
	"encoding/binary"
	"github.com/btcsuite/btcd/wire"
	btc "github.com/btcsuite/btcutil"
	"bytes"
	"strconv"
	"math/rand"
	"github.com/btcsuite/btcd/chaincfg"
	"fmt"
	"reflect"
	"strings"
	zmq "github.com/pebbe/zmq4"
	"time"
)

type Server struct {
	Url       string
	PublicKey string
}

type LibbitcoinClient struct {
	*ClientBase
	ServerList       []Server
	ConnectedServer  Server
	Params           chaincfg.Params
}

func NewLibbitcoinClient(servers []Server, params chaincfg.Params) *LibbitcoinClient {
	r := rand.Intn(len(servers))
	cb := NewClientBase(servers[r].Url, servers[r].PublicKey)
	client := LibbitcoinClient{
		ClientBase: cb,
		ServerList: servers,
		ConnectedServer: servers[r],
		Params: params,
	}
	go client.ListenHeartbeat(9092)
	return &client
}

func (l *LibbitcoinClient) ListenHeartbeat(port int) {
	i := strings.LastIndex(l.ConnectedServer.Url, ":")
	heartbeatUrl := l.ConnectedServer.Url[:i] + ":" + strconv.Itoa(port)
	c := make(chan Response)
	makeSocket := func() *ZMQSocket {
		s := NewSocket(c, zmq.SUB)
		s.Connect(heartbeatUrl, "")
		return s
	}
	s := makeSocket()

	timeout := func(){
		s.Close()
		fmt.Println("Server heartbeat timeout")
		// Rotate connected server here
		s = makeSocket()
	}
	ticker := time.NewTicker(10 * time.Second)
	go func() {
		for {
			select {
			case <- c:
				fmt.Println("heartbeat")
				ticker.Stop()
				ticker = time.NewTicker(10 * time.Second)
			case <- ticker.C:
				timeout()
			}
		}
	}()
}

func (l *LibbitcoinClient) FetchHistory2(address btc.Address, fromHeight uint32, callback func(interface{})) {
	hash160 := address.ScriptAddress()
	var netID byte
	height := make([]byte, 4)
	binary.LittleEndian.PutUint32(height, fromHeight)
	address.ScriptAddress()

	switch reflect.TypeOf(address).String() {

	case "btc.AddressPubKeyHash":
		if l.Params.Name == chaincfg.MainNetParams.Name {
			netID = byte(0)
		} else {
			netID = byte(111)
		}
	case "btc.AddressScriptHash":
		if l.Params.Name == chaincfg.MainNetParams.Name {
			netID = byte(5)
		} else {
			netID = byte(196)
		}
	}

	req := []byte{}
	req = append(req, netID)
	req = append(req, hash160...)
	req = append(req, height...)
	l.SendCommand("address.fetch_history2", req, callback)
}

func (l *LibbitcoinClient) FetchLastHeight(callback func(interface{})){
	l.SendCommand("blockchain.fetch_last_height", []byte{}, callback)
}

func ParseResponse(command string, data []byte, callback func(interface{})) {
	switch command {
	case "address.fetch_history2":
		numRows := (len(data)-4)/49
		buff := bytes.NewBuffer(data)
		buff.Next(4)
		rows := []FetchHistory2Row{}
		for i:=0; i<numRows; i++{
			r := FetchHistory2Row{}
			spendByte := buff.Next(1)
			spendBool, _ := strconv.ParseBool(string(spendByte))
			r.IsSpend = spendBool
			lehash := buff.Next(32)
			sh, _:= wire.NewShaHash(lehash)
			r.TxHash = sh.String()
			indexBytes := buff.Next(4)
			r.Index = binary.LittleEndian.Uint32(indexBytes)
			heightBytes := buff.Next(4)
			r.Height = binary.LittleEndian.Uint32(heightBytes)
			valueBytes := buff.Next(8)
			r.Value = binary.LittleEndian.Uint64(valueBytes)
			rows = append(rows, r)
		}
		callback(rows)
	case "blockchain.fetch_last_height":
		buff := bytes.NewBuffer(data)
		buff.Next(4)
		heightBytes := buff.Next(4)
		height := binary.LittleEndian.Uint32(heightBytes)
		callback(height)
	}
}