package types

import (
	"blocker/crypto"
	"blocker/proto"
	"crypto/sha256"

	pb "google.golang.org/protobuf/proto"
)

func VerifyBlock(b *proto.Block) bool {
	if len(b.PublicKey) != crypto.PubKeyLen {
		return false
	}
	if len(b.Signature) != crypto.SigLen {
		return false
	}
	hash := HashBlock(b)
	sig := crypto.SignatureFromBytes(b.Signature)
	pubKey := crypto.PublicKeyFromBytes(b.PublicKey)
	return sig.Verify(pubKey, hash)
}

func SignBlock(pk *crypto.PrivateKey, block *proto.Block) *crypto.Signature {
	hash := HashBlock(block)
	sig := pk.Sign(hash)
	block.PublicKey = pk.Public().Bytes()
	block.Signature = sig.Bytes()
	return sig

}

func HashBlock(block *proto.Block) []byte {
	return HashHeader(block.Header)
}

func HashHeader(header *proto.Header) []byte {
	b, err := pb.Marshal(header)
	if err != nil {
		panic(err)
	}
	hash := sha256.Sum256(b)
	return hash[:]
}
