package main

import (
	"blocker/node"
	"blocker/proto"
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	node := node.NewNode()
	fmt.Println("Start")

	go func() {
		for {
			time.Sleep(time.Second * 2)
			makeTransaction()

		}
	}()
	go func() {
		for {
			time.Sleep(time.Second * 5)
			makeTransaction()

		}
	}()
	err := node.Start(":4000")
	if err != nil {
		panic(err)
	}
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
