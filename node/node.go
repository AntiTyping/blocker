package node

import (
	"blocker/proto"
	"context"
	"fmt"
	"net"
	"os"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	peer "google.golang.org/grpc/peer"
)

type Node struct {
	version    string
	listenAddr string
	logger     *zap.SugaredLogger

	peerLock sync.RWMutex
	peers    map[proto.NodeClient]*proto.Version

	proto.UnimplementedNodeServer
}

func NewLogger() *zap.SugaredLogger {
	pe := zap.NewDevelopmentEncoderConfig()
	pe.EncodeTime = zapcore.ISO8601TimeEncoder
	pe.TimeKey = ""
	consoleEncoder := zapcore.NewConsoleEncoder(pe)
	level := zap.DebugLevel
	core := zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), level)
	logger := zap.New(core)
	defer logger.Sync()
	sugar := logger.Sugar()
	return sugar
}

func NewNode() *Node {
	return &Node{
		peers:   make(map[proto.NodeClient]*proto.Version),
		version: "blocker-1",
		logger:  NewLogger(),
	}
}

func (n *Node) addPeer(c proto.NodeClient, v *proto.Version) proto.NodeClient {
	n.logger.Infow("*** addPeer")
	n.peerLock.Lock()
	defer n.peerLock.Unlock()
	if _, ok := n.peers[c]; ok {
		return nil
	}
	n.peers[c] = v

	n.logger.Debugf("node %s added peer (%s): version %s height %d", n.listenAddr, v.ListenAddr, v.Version, v.Height)
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

func (n *Node) BootstrapNetwork(addrs []string) error {
	n.logger.Infow("*** Bootstrap node %s bootstrap network %s", n.listenAddr, addrs)
	for _, addr := range addrs {
		c, err := makeNodeClietn(addr)
		if err != nil {
			return err
		}
		v, err := c.Handshake(context.Background(), n.getVersion())
		if err != nil {
			fmt.Println("handshake error ", addrs, err)
			continue
		}

		n.addPeer(c, v)
	}
	return nil
}

func (n *Node) getVersion() *proto.Version {
	v := &proto.Version{
		Version:    n.version,
		Height:     1,
		ListenAddr: n.listenAddr,
	}
	return v
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
	n.logger.Infof("node running on port: ", listenAddr)
	return grpcSerer.Serve(ln)
}

func (n *Node) HandleTransaction(ctx context.Context, tx *proto.Transaction) (*proto.Ack, error) {
	peer, _ := peer.FromContext(ctx)
	fmt.Println("Received transaction from: ", peer)
	return &proto.Ack{}, nil
}

func (n *Node) Handshake(ctx context.Context, v *proto.Version) (*proto.Version, error) {
	p, _ := peer.FromContext(ctx)
	c, err := makeNodeClietn(v.ListenAddr)
	if err != nil {
		return nil, err
	}
	n.logger.Infof("node %s received hanshake from %s: %+v", n.listenAddr, p, v)

	n.addPeer(c, v)

	return n.getVersion(), nil
}

func makeNodeClietn(listenerAddr string) (proto.NodeClient, error) {
	client, err := grpc.Dial(listenerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	c := proto.NewNodeClient(client)
	return c, nil
}
