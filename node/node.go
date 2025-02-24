package node

import (
	"blocker/proto"
	"context"
	"fmt"
	"maps"
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

func (n *Node) Start(listenAddr string, bootstrapNodes []string) error {
	n.listenAddr = listenAddr
	opts := []grpc.ServerOption{}
	grpcSerer := grpc.NewServer(opts...)
	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return err
	}
	proto.RegisterNodeServer(grpcSerer, n)
	n.logger.Infof("[%s] node running on port: %s", listenAddr, listenAddr)

	if len(bootstrapNodes) > 0 {
		go n.bootstrapNetwork(bootstrapNodes)
	}

	return grpcSerer.Serve(ln)
}

func (n *Node) HandleTransaction(ctx context.Context, tx *proto.Transaction) (*proto.Ack, error) {
	peer, _ := peer.FromContext(ctx)
	fmt.Println("Received transaction from: ", peer)
	return &proto.Ack{}, nil
}

func (n *Node) Handshake(ctx context.Context, v *proto.Version) (*proto.Version, error) {
	n.logger.Infof("[%s] *** Hanshake from %s", n.listenAddr, v.ListenAddr)
	p, _ := peer.FromContext(ctx)
	c, err := makeNodeClietn(v.ListenAddr)
	if err != nil {
		return nil, err
	}
	n.logger.Infof("[%s] received hanshake from %s: %+v with peers %s", n.listenAddr, v.ListenAddr, p, v.PeerList)

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

func (n *Node) addPeer(c proto.NodeClient, v *proto.Version) proto.NodeClient {
	n.logger.Infof("[%s] *** addPeer %s", n.listenAddr, v)

	n.peerLock.Lock()
	defer n.peerLock.Unlock()
	if _, ok := n.peers[c]; ok {
		return nil
	}

	// handle the logic for the decision
	n.peers[c] = v

	n.logger.Debugf("[%s] node %s added peer %s: version %s height %d", n.listenAddr, n.listenAddr, v.ListenAddr, v.Version, v.Height)

	//for _, addr := range v.PeerList {
	//	n.logger.Infof("[%s] looking at peer from peer list  %s", n.listenAddr, addr)
	//	if addr != n.listenAddr {
	//		n.logger.Infof("[%s] need to connect with %s", n.listenAddr, addr)
	//		c, v, err := n.dialRemote(addr)
	//		if err != nil {
	//			n.logger.Infof("[%s] unable to peer with %s", n.listenAddr)
	//			continue
	//		}
	//		n.peers[*c] = v
	//		n.logger.Infof("[%s] added  %s", n.listenAddr, addr)
	//	}
	//}
	if len(v.PeerList) > 0 {
		go n.bootstrapNetwork(v.PeerList)
	}

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

func (n *Node) bootstrapNetwork(addrs []string) error {
	n.logger.Infof("[%s] Bootstrap node bootstrap network %v", n.listenAddr, addrs)
	for _, addr := range addrs {
		if !n.canConnect(addr) {
			continue
		}
		c, v, err := n.dialRemote(addr)
		if err != nil {
			return err
		}
		n.logger.Infof("[%s] received peer list %s", n.listenAddr, v.PeerList)

		n.addPeer(*c, v)
	}
	return nil
}

func (n *Node) getVersion() *proto.Version {
	v := &proto.Version{
		Version:    n.version,
		Height:     1,
		ListenAddr: n.listenAddr,
		PeerList:   n.getPeerList(),
	}
	return v
}

func (n *Node) canConnect(addr string) bool {
	if addr == n.listenAddr {
		return false
	}

	connecedPeers := n.getPeerList()
	for _, connecedPeer := range connecedPeers {
		if addr == connecedPeer {
			return false
		}
	}
	return true
}

func (n *Node) getPeerList() []string {
	n.peerLock.RLock()
	defer n.peerLock.RUnlock()
	peerList := make([]string, len(n.peers))
	i := 0
	for v := range maps.Values(n.peers) {
		peerList[i] = v.ListenAddr
		i++
	}
	return peerList
}

func (n *Node) dialRemote(addr string) (*proto.NodeClient, *proto.Version, error) {
	n.logger.Infof("[%s] *** dialRemote address %s", n.listenAddr, addr)
	c, err := makeNodeClietn(addr)
	if err != nil {
		return nil, nil, err
	}
	v, err := c.Handshake(context.Background(), n.getVersion())
	if err != nil {
		fmt.Println("handshake error ", addr, err)
		return &c, v, err
	}
	return &c, v, nil
}
