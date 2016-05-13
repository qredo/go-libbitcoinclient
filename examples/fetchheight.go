package main

import (
	"fmt"
	"github.com/btcsuite/btcd/chaincfg"
	libbitcoin "github.com/OpenBazaar/go-libbitcoinclient"
)

func main() {
	servers := []libbitcoin.Server{
		libbitcoin.Server{
			Url:"tcp://libbitcoin1.openbazaar.org:9091",
			PublicKey:"",
		},
	}
	client := libbitcoin.NewLibbitcoinClient(servers, chaincfg.MainNetParams)
	client.FetchLastHeight(func(i interface{}){
		fmt.Println(i.(uint32))
	})
	for {}
}