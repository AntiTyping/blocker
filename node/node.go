package node

import (
	"blocker/proto"
	"context"
	"fmt"
	"net"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	peer "google.golang.org/grpc/peer"
)

type Node struct {
	version    string
	listenAddr string

	peerLock sync.RWMutex
	peers    map[proto.NodeClient]*proto.Version

	proto.UnimplementedNodeServer
}

func NewNode() *Node {
	return &Node{
		peers:   make(map[proto.NodeClient]*proto.Version),
		version: "blocker-1",
	}
}

func (n *Node) addPeer(c proto.NodeClient, v *proto.Version) proto.NodeClient {
	n.peerLock.Lock()
	defer n.peerLock.Unlock()
	if _, ok := n.peers[c]; ok {
		return nil
	}
	n.peers[c] = v

	fmt.Printf("new peer connected (%s): height %d\n", v.ListenAddr, v.Height)
	return c
}

func (n *Node) removePeer(c proto.NodeClient) bool {
	n.peerLock.Lock()
	defer n.peerLock.Unlock()
	if _, ok := n.peers[c]; !ok {
		return false
	}

	delete(n.peers, c)
	return true
}

func (n *Node) deletePeer(c proto.NodeClient) {

}

func (n *Node) Start(listenAddr string) error {
	n.listenAddr = listenAddr
	opts := []grpc.ServerOption{}
	grpcSerer := grpc.NewServer(opts...)
	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return err
	}
	proto.RegisterNodeServer(grpcSerer, n)
	fmt.Println("node running on port: ", ":4000")
	return grpcSerer.Serve(ln)
}

func (n *Node) HandleTransaction(ctx context.Context, tx *proto.Transaction) (*proto.Ack, error) {
	peer, _ := peer.FromContext(ctx)
	fmt.Println("Received transaction from: ", peer)
	return &proto.Ack{}, nil
}

func (n *Node) Handshake(ctx context.Context, v *proto.Version) (*proto.Version, error) {
	ourVersion := &proto.Version{
		Version:    n.version,
		Height:     100,
		ListenAddr: n.listenAddr,
	}

	p, _ := peer.FromContext(ctx)
	c, err := makeNodeClietn(v.ListenAddr)
	if err != nil {
		return nil, err
	}

	n.addPeer(c, v)

	fmt.Printf("received version from %s: %+v\n", p, v)
	return ourVersion, nil
}

func makeNodeClietn(listenerAddr string) (proto.NodeClient, error) {
	client, err := grpc.Dial(listenerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	c := proto.NewNodeClient(client)
	return c, nil
}
