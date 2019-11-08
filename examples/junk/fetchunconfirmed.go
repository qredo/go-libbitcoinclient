package main

import (
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

	tx := "2d3024e7d75d4f12c4b879916fa0ffeca7e3d3d2885a789841542888304463a2"
	client.FetchUnconfirmedTransaction(tx, func(i interface{}, err error) {
		if err != nil {
			fmt.Println(err.Error())

		} else {
			fmt.Println(i.(btcutil.Tx))
		}
		wg.Done()
	})
	wg.Wait()
}
