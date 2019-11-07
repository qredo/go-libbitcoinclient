package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"sync"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
	libbitcoin "github.com/qredo/go-libbitcoinclient"
)

func main() {
	var wg sync.WaitGroup
	wg.Add(1)
	servers := []libbitcoin.Server{
		libbitcoin.Server{
			Url:       "tcp://mainnet.libbitcoin.net:9091",
			PublicKey: "",
		},
	}
	client := libbitcoin.NewLibbitcoinClient(servers, &chaincfg.MainNetParams)

	tx := "61222090ac412e1b77793f03b68b2551710b51949edb38d6f516e369eb499ea4"
	client.FetchTransaction(tx, func(i interface{}, err error) {
		if err != nil {
			fmt.Println(err.Error())

		} else {
			tx := i.(*btcutil.Tx)
			output := new(bytes.Buffer)
			tx.MsgTx().Serialize(output)
			fmt.Println(hex.EncodeToString(output.Bytes()))
		}
		wg.Done()
	})
	wg.Wait()
}
