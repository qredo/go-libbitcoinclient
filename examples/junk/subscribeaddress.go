package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"sync"

	"github.com/btcsuite/btcd/chaincfg"
	btc "github.com/btcsuite/btcutil"
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
	addr, _ := btc.DecodeAddress("1dice8EMZmqKvrGE4Qc9bUFf9PX3xaYDp", &chaincfg.MainNetParams)
	client.SubscribeAddress(addr, func(i interface{}) {
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
