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
	Params           *chaincfg.Params
	subscriptions    map[string]func(interface{}, error)
}

var client *LibbitcoinClient

func NewLibbitcoinClient(servers []Server, params *chaincfg.Params) *LibbitcoinClient {
	r := rand.Intn(len(servers))
	cb := NewClientBase(servers[r].Url, servers[r].PublicKey)
	subs := make(map[string]func(interface{}, error))
	c := LibbitcoinClient{
		ClientBase: cb,
		ServerList: servers,
		ConnectedServer: servers[r],
		Params: params,
		subscriptions: subs,
	}
	client = &c
	go c.ListenHeartbeat(9092)
	return &c
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

func (l *LibbitcoinClient) FetchHistory2(address btc.Address, fromHeight uint32, callback func(interface{}, error)) {
	hash160 := address.ScriptAddress()
	var netID byte
	height := make([]byte, 4)
	binary.LittleEndian.PutUint32(height, fromHeight)
	address.ScriptAddress()

	switch reflect.TypeOf(address).String() {

	case "*btcutil.AddressPubKeyHash":
		if l.Params.Name == chaincfg.MainNetParams.Name {
			netID = byte(0)
		} else {
			netID = byte(111)
		}
	case "*btcutil.AddressScriptHash":
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

func (l *LibbitcoinClient) FetchLastHeight(callback func(interface{}, error)){
	l.SendCommand("blockchain.fetch_last_height", []byte{}, callback)
}

func (l *LibbitcoinClient) FetchTransaction(txid string, callback func(interface{}, error)){
	b, _ := wire.NewShaHashFromStr(txid)
	l.SendCommand("blockchain.fetch_transaction", b.Bytes(), callback)
}

func (l *LibbitcoinClient) FetchUnconfirmedTransaction(txid string, callback func(interface{}, error)){
	b, _ := wire.NewShaHashFromStr(txid)
	l.SendCommand("transaction_pool.fetch_transaction", b.Bytes(), callback)
}

func (l *LibbitcoinClient) SubscribeAddress(address btc.Address, callback func(interface{}, error)) {
	req := []byte{}
	req = append(req, byte(0))
	req = append(req, byte(160))
	req = append(req, address.ScriptAddress()...)
	l.SendCommand("address.subscribe", req, nil)
	l.subscriptions[address.String()] = callback
}

func ParseResponse(command string, data []byte, callback func(interface{}, error)) {
	switch command {
	case "address.fetch_history2":
		numRows := (len(data)-4)/49
		buff := bytes.NewBuffer(data)
		err := ParseError(buff.Next(4))
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
		callback(rows, err)
	case "blockchain.fetch_last_height":
		height := binary.LittleEndian.Uint32(data[4:])
		callback(height, ParseError(data[:4]))
	case "blockchain.fetch_transaction":
		txn, _ := btc.NewTxFromBytes(data[4:])
		callback(txn, ParseError(data[:4]))
	case "transaction_pool.fetch_transaction":
		txn, _ := btc.NewTxFromBytes(data[4:])
		callback(txn, ParseError(data[:4]))
	case "address.update":
		buff := bytes.NewBuffer(data)
		addressVersion := buff.Next(1)[0]
		addressHash160 := buff.Next(20)
		height := buff.Next(4)
		block := buff.Next(32)
		tx := buff.Bytes()

		var addr btc.Address
		if addressVersion == byte(111) || addressVersion == byte(0) {
			a, _ := btc.NewAddressPubKeyHash(addressHash160, client.Params)
			addr = a
		} else if addressVersion == byte(5) || addressVersion == byte(196) {
			a, _ := btc.NewAddressScriptHashFromHash(addressHash160, client.Params)
			addr = a
		}
		bl, _ := wire.NewShaHash(block)
		txn, _ := btc.NewTxFromBytes(tx)

		resp := SubscribeResp{
			Address: addr.String(),
			Height: binary.LittleEndian.Uint32(height),
			Block: bl.String(),
			Tx: *txn,
		}
		client.subscriptions[addr.String()](resp, nil)
	}
}