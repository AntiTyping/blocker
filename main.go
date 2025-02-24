package main

import (
	"blocker/node"
	"blocker/proto"
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	go func() { makeNode(":5001", []string{}) }()
	time.Sleep(time.Second * 1)
	go func() { makeNode(":5002", []string{":5001"}) }()

	time.Sleep(4 * time.Second)
	go func() { makeNode(":6000", []string{":5002"}) }()

	time.Sleep(time.Second * 20)
}

func makeNode(listenAddr string, bootstrapNodes []string) *node.Node {
	n := node.NewNode()
	go n.Start(listenAddr, bootstrapNodes)
	return n
}

func makeTransaction() {
	client, err := grpc.Dial(":4000", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	c := proto.NewNodeClient(client)

	tx := &proto.Version{
		Version:    "blocker-1",
		Height:     0,
		ListenAddr: ":123",
	}

	_, err = c.Handshake(context.Background(), tx)
	if err != nil {
		panic(err)
	}
}
