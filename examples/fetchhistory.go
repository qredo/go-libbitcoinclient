package main

import (
	"fmt"
	"github.com/btcsuite/btcd/chaincfg"
	btc "github.com/btcsuite/btcutil"
	libbitcoin "github.com/OpenBazaar/go-libbitcoinclient"
	"time"
)

func main() {
	addr, _ := btc.DecodeAddress("mrhqn9X8A121nn2AZCwqSdHcdQqttKKG45", &chaincfg.TestNet3Params)
	servers := []libbitcoin.Server{
		libbitcoin.Server{
			Url:"tcp://libbitcoin2.openbazaar.org:9091",
			PublicKey:"baihZB[vT(dcVCwkhYLAzah<t2gJ>{3@k?+>T&^3",
		},
	}
	client := libbitcoin.NewLibbitcoinClient(servers, &chaincfg.TestNet3Params)
	client.FetchHistory2(addr, uint32(0), func(i interface{}, err error){
		for _, response := range(i.([]libbitcoin.FetchHistory2Row)){
			fmt.Printf("Txid: %s\n", response.TxHash)
			fmt.Printf("Index: %d\n", response.Index)
			fmt.Printf("Is Spend: %t\n", response.IsSpend)
			fmt.Printf("Chain Height: %d\n", response.Height)
			fmt.Printf("Value (satoshis): %d\n", response.Value)
			fmt.Println()
		}
	})
	time.Sleep(10 *time.Second)
}