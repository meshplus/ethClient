package main

import (
	"fmt"
	"github.com/meshplus/ethClient/eth"
	"log"
	"time"
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
	//deploy, _, err := client.Deploy("http://127.0.0.1:8881", "example/transfer.sol", "0xD3880ea40670eD51C3e3C0ea089fDbDc9e3FBBb4", true)
	//if err != nil {
	//	fmt.Println("deploy err", err)
	//}
	//fmt.Println("deploy", deploy)
	//
	//fmt.Println(client)
	// Send 1 eth
	txid, err := client.EthSendTransaction(&eth.T{
		To:    "0xcfa202c4268749fbb5136f2b68f7402984ed444b",
		Value: eth.Eth1(),
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(txid)
	time.Sleep(1 * time.Second)
	receipt, err := client.EthGetTransactionReceipt(txid)
	if err != nil {
		fmt.Println("error", err)
		return
	}
	fmt.Println("recepit", receipt)
	//contract, err := client.InvokeEthContract("example/broker.abi", "0xD3880ea40670eD51C3e3C0ea089fDbDc9e3FBBb4", "audit", "0x668a209Dc6562707469374B8235e37b8eC25db08^1")
	//if err != nil {
	//	fmt.Println(err)
	//}
	//fmt.Println(contract)
}
