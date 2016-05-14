package main

import (
	"fmt"
	"github.com/btcsuite/btcd/chaincfg"
	libbitcoin "github.com/OpenBazaar/go-libbitcoinclient"
	"time"
	"github.com/btcsuite/btcutil"
)

func main() {
	servers := []libbitcoin.Server{
		libbitcoin.Server{
			Url:"tcp://libbitcoin3.openbazaar.org:9091",
			PublicKey:"",
		},
	}
	client := libbitcoin.NewLibbitcoinClient(servers, &chaincfg.MainNetParams)

	tx := "d26600672e219914c37aca78850b17e01bbeab3252e6239da0377bcb63e3e119"
	client.FetchTransaction(tx, func(i interface{}, err error){
		fmt.Printf(i.(btcutil.Tx))
	})
	time.Sleep(10 *time.Second)
}