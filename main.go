package main

import (
	"blocker/crypto"
	"blocker/node"
	"blocker/proto"
	"blocker/util"
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	go func() { makeNode(":5001", []string{}, true) }()
	time.Sleep(time.Second * 1)
	go func() { makeNode(":5002", []string{":5001"}, false) }()

	time.Sleep(4 * time.Second)
	go func() { makeNode(":6000", []string{":5002"}, false) }()

	for {
		time.Sleep(time.Millisecond * 200)
		makeTransaction()
	}
}

func makeNode(listenAddr string, bootstrapNodes []string, validator bool) *node.Node {
	cfg := node.ServerConfig{
		Version:    "blocker-1",
		ListenAddr: listenAddr,
	}

	if validator {
		cfg.PrivateKey = crypto.GeneratePrivateKey()
	}

	n := node.NewNode(cfg)
	go n.Start(listenAddr, bootstrapNodes)
	return n
}

func makeTransaction() {
	client, err := grpc.Dial(":5001", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	c := proto.NewNodeClient(client)

	privKey := crypto.GeneratePrivateKey()

	tx := &proto.Transaction{
		Version: 1,
		Inputs: []*proto.TxInput{
			{
				PrevTxHash:   util.RandomHash(),
				PrevOutIndex: 0,
				PublicKey:    privKey.Public().Bytes(),
			},
		},
		Outputs: []*proto.TxOutput{
			{
				Amount:    99,
				ToAddress: privKey.Public().Bytes(),
			},
		},
	}

	_, err = c.HandleTransaction(context.Background(), tx)
	if err != nil {
		panic(err)
	}
}
