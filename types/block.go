package types

import (
	"blocker/crypto"
	"blocker/proto"
	"crypto/sha256"

	pb "google.golang.org/protobuf/proto"
)

func SignBlock(pk *crypto.PrivateKey, block *proto.Block) *crypto.Signature {
	return pk.Sign(HashBlock(block))

}

func HashBlock(block *proto.Block) []byte {
	b, err := pb.Marshal(block)
	if err != nil {
		panic(err)
	}

	hash := sha256.Sum256(b)
	return hash[:]
}
