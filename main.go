package main

import (
	"ethClient/eth"
	"fmt"
	"log"
)

func main() {
	client, err := eth.New("http://127.0.0.1:8881", "./config/")
	if err != nil {
		log.Fatal(err)
	}
	//fmt.Println(client)
	//compile, err := client.Compile("./example/broker.sol", true)
	//if err != nil {
	//	fmt.Println("err", err)
	//}
	//fmt.Println(compile)
	//deploy, _, err := client.Deploy("http://127.0.0.1:8881", "example/storage.sol", "", true)
	//if err != nil {
	//	fmt.Println("deploy err", err)
	//}
	//fmt.Println("deploy", deploy)
	//
	//fmt.Println(client)
	// Send 1 eth
	txid, err := client.EthSendTransaction(&eth.T{
		From:  "0x6247cf0412c6462da2a51d05139e2a3c6c630f0a",
		To:    "0xcfa202c4268749fbb5136f2b68f7402984ed444b",
		Value: eth.Eth1(),
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(txid)
}
