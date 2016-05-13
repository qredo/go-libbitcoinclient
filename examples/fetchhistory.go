package main

import (
	"fmt"
	"github.com/btcsuite/btcd/chaincfg"
	btc "github.com/btcsuite/btcutil"
	libbitcoin "github.com/OpenBazaar/go-libbitcoinclient"
)

func main() {
	addr, _ := btc.DecodeAddress("3GprJWkxHx3v9qbvRALcquXPHyNjqSTjvy", &chaincfg.MainNetParams)
	servers := []libbitcoin.Server{
		libbitcoin.Server{
			Url:"tcp://libbitcoin1.openbazaar.org:9091",
			PublicKey:"",
		},
	}
	client := libbitcoin.NewLibbitcoinClient(servers, chaincfg.MainNetParams)
	client.FetchHistory2(addr, uint32(0), func(i interface{}){
		for _, response := range(i.([]libbitcoin.FetchHistory2Row)){
			fmt.Printf("Txid: %s\n", response.TxHash)
			fmt.Printf("Index: %d\n", response.Index)
			fmt.Printf("Is Spend: %t\n", response.IsSpend)
			fmt.Printf("Chain Height: %d\n", response.Height)
			fmt.Printf("Value (satoshis): %d\n", response.Value)
			fmt.Println()
		}
	})
	for {}
}