package main

import (
	"fmt"
	"sync"
	"github.com/btcsuite/btcd/chaincfg"
	btc "github.com/btcsuite/btcutil"
	libbitcoin "github.com/qredo/go-libbitcoinclient"
)

func main() {
	var wg sync.WaitGroup
	wg.Add(1)
	addr, _ := btc.DecodeAddress("1Ec9S8KSw4UXXhqkoG3ZD31yjtModULKGg", &chaincfg.MainNetParams)
	servers := []libbitcoin.Server{
		libbitcoin.Server{
			Url:"tcp://mainnet.libbitcoin.net:9091",
			PublicKey:"",
		},
	}
	client := libbitcoin.NewLibbitcoinClient(servers, &chaincfg.MainNetParams)
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