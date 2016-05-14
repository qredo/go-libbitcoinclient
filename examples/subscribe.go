package main

import (
	"fmt"
	"github.com/btcsuite/btcd/chaincfg"
	btc "github.com/btcsuite/btcutil"
	libbitcoin "github.com/OpenBazaar/go-libbitcoinclient"
	"time"
)

func main() {
	servers := []libbitcoin.Server{
		libbitcoin.Server{
			Url:"tcp://libbitcoin2.openbazaar.org:9091",
			PublicKey:"baihZB[vT(dcVCwkhYLAzah<t2gJ>{3@k?+>T&^3",
		},
	}
	client := libbitcoin.NewLibbitcoinClient(servers, &chaincfg.TestNet3Params)
	addr, _ := btc.DecodeAddress("mrhqn9X8A121nn2AZCwqSdHcdQqttKKG45", &chaincfg.TestNet3Params)
	client.SubscribeAddress(addr, func(i interface{}){
		resp := i.(libbitcoin.SubscribeResp)
		fmt.Println(resp.Address)
		fmt.Println(resp.Height)
		fmt.Println(resp.Block)
		fmt.Println(resp.Tx)
		fmt.Println()
	})
	time.Sleep(60 *time.Second)
}