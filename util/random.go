package util

import (
	"blocker/proto"
	crand "crypto/rand"
	"io"
	"math/rand"
	"time"
)

func RandomHash() []byte {
	hash := make([]byte, 32)
	io.ReadFull(crand.Reader, hash)
	return hash
}

func RandomBlock() *proto.Block {
	header := &proto.Header{
		Version:      1,
		Height:       int32(rand.Intn(1000)),
		PreviousHash: RandomHash(),
		RootHash:     RandomHash(),
		Timestamp:    time.Now().UnixNano(),
	}
	return &proto.Block{
		Header: header,
	}
}
