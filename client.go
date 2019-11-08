package libbitcoin

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/btcsuite/btcd/chaincfg"

	btc "github.com/btcsuite/btcutil"
	"github.com/op/go-logging"
	zmq "github.com/pebbe/zmq4"
)

type ShaHash [32]byte

const HashSize = 32
const MaxHashStringSize = HashSize * 2

var ErrHashStrSize = fmt.Errorf("max hash string length is %v bytes", MaxHashStringSize)

var log = logging.MustGetLogger("main")

const (
	HeartbeatPort    = 9092
	BlockPublishPort = 9093
)

type Server struct {
	Url       string
	PublicKey string
}

type LibbitcoinClient struct {
	*ClientBase
	ServerList     []Server
	ServerIndex    int
	Params         *chaincfg.Params
	subscriptions  map[string]subscription
	connectionTime time.Time
	lock           *sync.Mutex
}

type subscription struct {
	expiration time.Time
	callback   func(interface{})
}

func NewLibbitcoinClient(servers []Server, params *chaincfg.Params) *LibbitcoinClient {
	rb, _ := rand.Int(rand.Reader, big.NewInt(int64(len(servers))))
	r := int(rb.Int64())
	cb := NewClientBase(servers[r].Url, servers[r].PublicKey)
	subs := make(map[string]subscription)
	l := new(sync.Mutex)
	client := LibbitcoinClient{
		ClientBase:     cb,
		ServerList:     servers,
		ServerIndex:    r,
		Params:         params,
		subscriptions:  subs,
		connectionTime: time.Now(),
		lock:           l,
	}
	cb.parser = client.Parse
	cb.timeout = client.RotateServer
	go client.ListenHeartbeat()
	go client.renewSubscriptions()
	log.Infof("Libbitcoin client connected to %s\n", client.ServerList[client.ServerIndex].Url)
	return &client
}

func (l *LibbitcoinClient) RotateServer() {
	if time.Now().Sub(l.connectionTime) > time.Second*30 {
		currentUrl := l.ServerList[l.ServerIndex].Url
		l.ServerIndex = (l.ServerIndex + 1) % len(l.ServerList)
		l.ClientBase.socket.ChangeEndpoint(currentUrl, l.ServerList[l.ServerIndex].Url, l.ServerList[l.ServerIndex].PublicKey)
		l.lock.Lock()
		for k, v := range l.subscriptions {
			addr, _ := btc.DecodeAddress(k, l.Params)
			l.SubscribeAddress(addr, v.callback)
		}
		l.lock.Unlock()
		l.connectionTime = time.Now()
		log.Infof("Rotating libbitcoin server, using %s\n", l.ServerList[l.ServerIndex].Url)
	}
}

func (l *LibbitcoinClient) ListenHeartbeat() {
	i := strings.LastIndex(l.ServerList[l.ServerIndex].Url, ":")
	heartbeatUrl := l.ServerList[l.ServerIndex].Url[:i] + ":" + strconv.Itoa(HeartbeatPort)
	c := make(chan Response)
	s := NewSocket(c, zmq.SUB)
	s.Connect(heartbeatUrl, "")

	timeout := func() {
		log.Warningf("Libbitcoin server at %s timed out on heartbeat\n", l.ServerList[l.ServerIndex].Url)
		l.RotateServer()
		currentUrl := heartbeatUrl
		i := strings.LastIndex(l.ServerList[l.ServerIndex].Url, ":")
		heartbeatUrl = l.ServerList[l.ServerIndex].Url[:i] + ":" + strconv.Itoa(HeartbeatPort)
		s.ChangeEndpoint(currentUrl, heartbeatUrl, "")
	}
	ticker := time.NewTicker(10 * time.Second)
	for {
		select {
		case <-c:
			ticker.Stop()
			ticker = time.NewTicker(10 * time.Second)
		case <-ticker.C:
			timeout()
		}
	}
}

func (l *LibbitcoinClient) renewSubscriptions() {
	ticker := time.NewTicker(1 * time.Minute)
	func() {
		for {
			select {
			case <-ticker.C:
				l.lock.Lock()
				for k, v := range l.subscriptions {
					if v.expiration.After(time.Now()) {
						addr, _ := btc.DecodeAddress(k, l.Params)
						l.RenewSubscription(addr, v.callback)
					}
				}
				l.lock.Unlock()
			}
		}
	}()
}

func (l *LibbitcoinClient) FetchHistory2(address btc.Address, fromHeight uint32, callback func(interface{}, error)) {
	hash160 := address.ScriptAddress()
	//var netID byte
	height := make([]byte, 4)
	binary.LittleEndian.PutUint32(height, fromHeight)
	address.ScriptAddress()

	// switch reflect.TypeOf(address).String() {

	// case "*btcutil.AddressPubKeyHash":
	// 	if l.Params.Name == chaincfg.MainNetParams.Name {
	// 		netID = byte(0)
	// 	} else {
	// 		netID = byte(111)
	// 	}
	// case "*btcutil.AddressScriptHash":
	// 	if l.Params.Name == chaincfg.MainNetParams.Name {
	// 		netID = byte(5)
	// 	} else {
	// 		netID = byte(196)
	// 	}
	// }
	req := []byte{}
	//req = append(req, netID)
	req = append(req, hash160...)
	req = append(req, height...)
	go l.SendCommand("blockchain.fetch_history3", req, callback)
}

func (l *LibbitcoinClient) FetchHistory3(address btc.Address, fromHeight uint32, callback func(interface{}, error)) {
	hash160 := address.ScriptAddress()
	height := make([]byte, 4)
	binary.LittleEndian.PutUint32(height, fromHeight)
	req := []byte{}
	req = append(req, hash160...)
	req = append(req, height...)
	go l.SendCommand("blockchain.fetch_history3", req, callback)
}

func (l *LibbitcoinClient) FetchLastHeight(callback func(interface{}, error)) {
	go l.SendCommand("blockchain.fetch_last_height", []byte{}, callback)
}

func (l *LibbitcoinClient) FetchTransaction(txid string, callback func(interface{}, error)) {
	b, _ := NewShaHashFromStr(txid)
	go l.SendCommand("blockchain.fetch_transaction", b.Bytes(), callback)
}

func (l *LibbitcoinClient) FetchUnconfirmedTransaction(txid string, callback func(interface{}, error)) {
	b, _ := NewShaHashFromStr(txid)
	go l.SendCommand("transaction_pool.fetch_transaction", b.Bytes(), callback)
}

func (l *LibbitcoinClient) SubscribeAddress(address btc.Address, callback func(interface{})) {
	req := []byte{}
	req = append(req, byte(0))
	req = append(req, byte(160))
	req = append(req, address.ScriptAddress()...)
	go l.SendCommand("address.subscribe", req, nil)
	l.lock.Lock()
	l.subscriptions[address.String()] = subscription{
		expiration: time.Now().Add(24 * time.Hour),
		callback:   callback,
	}
	l.lock.Unlock()
}

func (l *LibbitcoinClient) UnsubscribeAddress(address btc.Address) {
	l.lock.Lock()
	_, ok := l.subscriptions[address.String()]
	if ok {
		delete(l.subscriptions, address.String())
	}
	l.lock.Unlock()
}

func (l *LibbitcoinClient) RenewSubscription(address btc.Address, callback func(interface{})) {
	req := []byte{}
	req = append(req, byte(0))
	req = append(req, byte(160))
	req = append(req, address.ScriptAddress()...)
	go l.SendCommand("address.renew", req, nil)
	l.lock.Lock()
	l.subscriptions[address.String()] = subscription{
		expiration: time.Now().Add(24 * time.Hour),
		callback:   callback,
	}
	l.lock.Unlock()
}

func (l *LibbitcoinClient) Broadcast(tx []byte, callback func(interface{}, error)) {
	go l.SendCommand("protocol.broadcast_transaction", tx, callback)
}

func (l *LibbitcoinClient) Validate(tx []byte, callback func(interface{}, error)) {
	go l.SendCommand("transaction_pool.validate", tx, nil)
}

func (l *LibbitcoinClient) Parse(command string, data []byte, callback func(interface{}, error)) {
	switch command {
	case "blockchain.fetch_history3":
		numRows := (len(data) - 4) / 49
		buff := bytes.NewBuffer(data)
		err := ParseError(buff.Next(4))
		rows := []FetchHistory2Resp{}
		for i := 0; i < numRows; i++ {
			r := FetchHistory2Resp{}
			spendByte := int(buff.Next(1)[0])
			var spendBool bool
			if spendByte == 0 {
				spendBool = false
			} else {
				spendBool = true
			}
			r.IsSpend = spendBool
			lehash := buff.Next(32)
			sh, _ := NewShaHash(lehash)
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
			a, _ := btc.NewAddressPubKeyHash(addressHash160, l.Params)
			addr = a
		} else if addressVersion == byte(5) || addressVersion == byte(196) {
			a, _ := btc.NewAddressScriptHashFromHash(addressHash160, l.Params)
			addr = a
		}
		bl, _ := NewShaHash(block)
		txn, _ := btc.NewTxFromBytes(tx)

		resp := SubscribeResp{
			Address: addr.String(),
			Height:  binary.LittleEndian.Uint32(height),
			Block:   bl.String(),
			Tx:      *txn,
		}
		l.lock.Lock()
		l.subscriptions[addr.String()].callback(resp)
		l.lock.Unlock()
	case "protocol.broadcast_transaction":
		callback(nil, ParseError(data[:4]))
	case "transaction_pool.validate":
		b := data[4:5]
		success, _ := strconv.ParseBool(string(b))
		callback(success, ParseError(data[:4]))
	}
}

func NewShaHashFromStr(hash string) (*ShaHash, error) {
	// Return error if hash string is too long.
	if len(hash) > MaxHashStringSize {
		return nil, ErrHashStrSize
	}

	// Hex decoder expects the hash to be a multiple of two.
	if len(hash)%2 != 0 {
		hash = "0" + hash
	}

	// Convert string hash to bytes.
	buf, err := hex.DecodeString(hash)
	if err != nil {
		return nil, err
	}

	// Un-reverse the decoded bytes, copying into in leading bytes of a
	// ShaHash.  There is no need to explicitly pad the result as any
	// missing (when len(buf) < HashSize) bytes from the decoded hex string
	// will remain zeros at the end of the ShaHash.
	var ret ShaHash
	blen := len(buf)
	mid := blen / 2
	if blen%2 != 0 {
		mid++
	}
	blen--
	for i, b := range buf[:mid] {
		ret[i], ret[blen-i] = buf[blen-i], b
	}
	return &ret, nil
}

func NewShaHash(newHash []byte) (*ShaHash, error) {
	var sh ShaHash
	err := sh.SetBytes(newHash)
	if err != nil {
		return nil, err
	}
	return &sh, err
}

func (hash *ShaHash) Bytes() []byte {
	newHash := make([]byte, HashSize)
	copy(newHash, hash[:])

	return newHash
}

// SetBytes sets the bytes which represent the hash.  An error is returned if
// the number of bytes passed in is not HashSize.
func (hash *ShaHash) SetBytes(newHash []byte) error {
	nhlen := len(newHash)
	if nhlen != HashSize {
		return fmt.Errorf("invalid sha length of %v, want %v", nhlen,
			HashSize)
	}
	copy(hash[:], newHash[0:HashSize])

	return nil
}

func (hash ShaHash) String() string {
	for i := 0; i < HashSize/2; i++ {
		hash[i], hash[HashSize-1-i] = hash[HashSize-1-i], hash[i]
	}
	return hex.EncodeToString(hash[:])
}
