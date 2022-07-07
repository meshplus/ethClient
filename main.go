package main

import (
	"fmt"
	"github.com/meshplus/ethClient/eth"
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
	deploy, _, err := client.Deploy("http://127.0.0.1:8881", "example/broker.sol", "1356^appchain1^[0xc7F999b83Af6DF9e67d0a37Ee7e900bF38b3D013,0x79a1215469FaB6f9c63c1816b45183AD3624bE34,0x97c8B516D19edBf575D72a172Af7F418BE498C37,0xc0Ff2e0b3189132D815b8eb325bE17285AC898f8]^3^[0x000f1a7a08ccc48e5d30f80850cf1cf283aa3abd]^1", true)
	if err != nil {
		fmt.Println("deploy err", err)
	}
	fmt.Println("deploy", deploy)
	//
	//fmt.Println(client)
	// Send 1 eth
	//txid, err := client.EthSendTransaction(&eth.T{
	//	To:    "0xcfa202c4268749fbb5136f2b68f7402984ed444b",
	//	Value: eth.Eth1(),
	//})
	//if err != nil {
	//	log.Fatal(err)
	//}
	//fmt.Println(txid)
	//time.Sleep(1 * time.Second)
	//receipt, err := client.EthGetTransactionReceipt(txid)
	//if err != nil {
	//	fmt.Println("error", err)
	//	return
	//}
	//fmt.Println("recepit", receipt)
	//contract, err := client.InvokeEthContract("example/broker.abi", "0xD3880ea40670eD51C3e3C0ea089fDbDc9e3FBBb4", "audit", "0x668a209Dc6562707469374B8235e37b8eC25db08^1")
	//if err != nil {
	//	fmt.Println(err)
	//}
	//fmt.Println(contract)
}
