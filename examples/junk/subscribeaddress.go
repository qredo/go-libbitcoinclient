package main

import (
	"fmt"
	"sync"
	"bytes"
	"encoding/hex"
	"github.com/btcsuite/btcd/chaincfg"
	btc "github.com/btcsuite/btcutil"
	libbitcoin "github.com/qredo/go-libbitcoinclient"
)

func main() {
	var wg sync.WaitGroup
	wg.Add(1)
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
		fmt.Printf("Address: %s\n", resp.Address)
		fmt.Printf("Height: %d\n", resp.Height)
		fmt.Printf("Block: %s\n", resp.Block)
		output := new(bytes.Buffer)
		resp.Tx.MsgTx().Serialize(output)
		fmt.Printf("Tx: %s\n", hex.EncodeToString(output.Bytes()))
		fmt.Println()
		wg.Done()
	})
	wg.Wait()
}