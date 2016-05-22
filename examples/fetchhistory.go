package main

import (
	"fmt"
	"sync"
	"github.com/btcsuite/btcd/chaincfg"
	btc "github.com/btcsuite/btcutil"
	libbitcoin "github.com/OpenBazaar/go-libbitcoinclient"
)

func main() {
	var wg sync.WaitGroup
	wg.Add(1)
	addr, _ := btc.DecodeAddress("2Mu1qcdDxfy7ebH2yUasPkM36r3qw3AEGEG", &chaincfg.TestNet3Params)
	servers := []libbitcoin.Server{
		libbitcoin.Server{
			Url:"tcp://libbitcoin2.openbazaar.org:9091",
			PublicKey:"baihZB[vT(dcVCwkhYLAzah<t2gJ>{3@k?+>T&^3",
		},
	}
	client := libbitcoin.NewLibbitcoinClient(servers, &chaincfg.TestNet3Params)
	fromHeight := uint32(0)
	client.FetchHistory2(addr, fromHeight, func(i interface{}, err error){
		for _, response := range(i.([]libbitcoin.FetchHistory2Resp)){
			fmt.Printf("Txid: %s\n", response.TxHash)
			fmt.Printf("Index: %d\n", response.Index)
			fmt.Printf("Is Spend: %t\n", response.IsSpend)
			fmt.Printf("Chain Height: %d\n", response.Height)
			fmt.Printf("Value (satoshis): %d\n", response.Value)
			fmt.Println()
		}
		wg.Done()
	})
	wg.Wait()
}