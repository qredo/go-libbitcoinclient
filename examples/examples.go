package main

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"sync"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
	libbitcoin "github.com/qredo/go-libbitcoinclient"
)

/*
	2 in 2 out 1P9Sa7Dn8EK4KkN6YGc6Wfc8FkHxL58Fk6
	543 TX with remaining balance 3CRZtw8oL4mL5kfRYfvwwQSZ5UTdh9MWoK
*/
var (
	add1      = "3PeEHZsh3ZMMsjFY4QNCT5W21MuA26ti4q"
	add2      = "13ejSKUxLT9yByyr1bsLNseLbx9H9tNj2d"
	add3      = "1DvERptSm5urPyAecGyL6gKdZedXUrr9H9"
	url       = "tcp://mainnet.libbitcoin.net:9091"
	txid1     = "16359c90b0329265aab333fa464f69853516038a01ade8997afff26950d20a19"
	publickey = ""
	params    = &chaincfg.MainNetParams
)

// var (
// 	add1      = "mm9hVHDZReh3E9HkHtqzyEP1EA9DfAwhTA"
// 	url       = "tcp://testnet1.libbitcoin.net:19091"
// 	txid1     = "36291c89d289f8f5e5eaf1b6744d86e24a43d271291f33693b6aab2adb954733"
// 	publickey = ""
// 	params    = &chaincfg.TestNet3Params
// )

func main() {
	client := makeClient()
	address := add1
	//Block Height
	print(blockHeight(client))

	//Single transaction
	tx, _ := fetchTX(client, txid1)
	fmt.Println(hex.EncodeToString(tx))

	//History for an address
	txs, _ := fetchHistory(client, address)
	for _, response := range txs {
		fmt.Printf("Txid: %s\n", response.TxHash)
		fmt.Printf("Index: %d\n", response.Index)
		fmt.Printf("Is Spend: %t\n", response.IsSpend)
		fmt.Printf("Chain Height: %d\n", response.Height)
		fmt.Printf("Value (satoshis): %d\n", response.Value)
		fmt.Println()
	}

	//Balance
	balance, lastTXHeight, _ := fetchBalance(client, address)
	print("Balance:", balance, "\n")
	print("lastTXHeight:", lastTXHeight, "\n")

}

func makeClient() *libbitcoin.LibbitcoinClient {
	servers := []libbitcoin.Server{
		libbitcoin.Server{
			Url:       url,
			PublicKey: "",
		},
	}
	return libbitcoin.NewLibbitcoinClient(servers, params)
}

func blockHeight(client *libbitcoin.LibbitcoinClient) (blockHeight uint32) {
	var wg sync.WaitGroup
	wg.Add(1)

	client.FetchLastHeight(func(i interface{}, err error) {
		blockHeight = i.(uint32)
		wg.Done()
	})
	wg.Wait()
	return
}

func fetchTX(client *libbitcoin.LibbitcoinClient, txid string) (tx []byte, err error) {
	var wg sync.WaitGroup
	wg.Add(1)

	output := new(bytes.Buffer)
	client.FetchTransaction(txid, func(i interface{}, err error) {
		if err != nil {
			fmt.Println(err.Error())

		} else {
			rawtx := i.(*btcutil.Tx)
			rawtx.MsgTx().Serialize(output)

		}
		wg.Done()
	})
	wg.Wait()
	return output.Bytes(), err
}

func fetchHistory(client *libbitcoin.LibbitcoinClient, address string) (tx []libbitcoin.FetchHistory2Resp, err error) {
	var wg sync.WaitGroup
	wg.Add(1)
	addr, _ := btcutil.DecodeAddress(address, params)
	fromHeight := uint32(0)
	client.FetchHistory3(addr, fromHeight, func(i interface{}, err error) {
		tx = i.([]libbitcoin.FetchHistory2Resp)
		wg.Done()
	})
	wg.Wait()
	return tx, err
}

func fetchBalance(client *libbitcoin.LibbitcoinClient, address string) (satoshis uint64, lastTXHeight uint32, err error) {
	txs, err := fetchHistory(client, address)
	if err != nil {
		return 0, 0, err
	}
	sub := make(map[uint64]uint64)
	//Pass one add all the input UTXOs
	for _, response := range txs {
		if response.IsSpend == false {
			satoshis += response.Value
			if response.Height > lastTXHeight {
				lastTXHeight = response.Height
			}
			cs := checksum(response.TxHash, response.Index)
			sub[cs] = response.Value
		}
	}
	//Pass 2 subtract all the macthing spends
	for _, response := range txs {
		if response.IsSpend == true {
			satoshis -= sub[response.Value]
		}
	}
	return satoshis, lastTXHeight, nil

}

func checksum(tx string, index uint32) uint64 {
	const mask uint64 = 0xffffffffffff8000
	const invmask uint32 = 0x00007FFF

	hash, _ := hex.DecodeString(tx[:40])
	//hashpart, _ := hex.DecodeString("36291c89d289f8f5e5eaf1b6744d86e24a43d271")
	//hashpart, _ := hex.DecodeString("71d2434ae2864d74b6f1eae5f5f889d2891c2936")

	upper := binary.LittleEndian.Uint64(reverse(hash)) & mask
	lower := index & invmask
	answer := upper | uint64(lower)
	return answer
}

func reverse(b []byte) []byte {
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}
	return b
}
