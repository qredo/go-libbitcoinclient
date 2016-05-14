package main

import (
	"fmt"
	"time"
	"bytes"
	"encoding/hex"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
	libbitcoin "github.com/OpenBazaar/go-libbitcoinclient"
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
		if err != nil {
			fmt.Println(err.Error())

		} else {
			tx := i.(*btcutil.Tx)
			output := new(bytes.Buffer)
			tx.MsgTx().Serialize(output)
			fmt.Println(hex.EncodeToString(output.Bytes()))
		}
	})
	time.Sleep(10 *time.Second)
}