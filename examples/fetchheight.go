package main

import (
	"fmt"
	"sync"
	"github.com/btcsuite/btcd/chaincfg"
	libbitcoin "github.com/OpenBazaar/go-libbitcoinclient"
)

func main() {
	var wg sync.WaitGroup
	wg.Add(1)
	servers := []libbitcoin.Server{
		libbitcoin.Server{
			Url:"tcp://libbitcoin5.openbazaar.org:9091",
			PublicKey:"",
		},
	}
	client := libbitcoin.NewLibbitcoinClient(servers, &chaincfg.MainNetParams)
	client.FetchLastHeight(func(i interface{}, err error){
		fmt.Println(i.(uint32))
		//wg.Done()
	})
	wg.Wait()
}