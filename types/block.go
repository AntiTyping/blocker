package types

import (
	"blocker/crypto"
	"blocker/proto"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/cbergoon/merkletree"
	pb "google.golang.org/protobuf/proto"
)

type TxHash struct {
	hash []byte
}

func NewTxHash(hash []byte) TxHash {
	return TxHash{hash}
}

func (h TxHash) CalculateHash() ([]byte, error) {
	return h.hash, nil
}

func (h TxHash) Equals(other merkletree.Content) (bool, error) {
	return bytes.Equal(h.hash, other.(TxHash).hash), nil
}

func VerifyBlock(b *proto.Block) (bool, error) {
	if len(b.Transactions) > 0 {
		if valid, err := VerifyRootHash(b); err != nil || !valid {
			if err != nil {
				return false, err
			}
			return false, fmt.Errorf("invalid root hash %s", hex.EncodeToString(b.Header.RootHash))
		}
	}
	if len(b.PublicKey) != crypto.PubKeyLen {
		return false, fmt.Errorf("wrong public key lenght")
	}
	if len(b.Signature) != crypto.SigLen {
		return false, fmt.Errorf("wrong signature length")
	}
	hash := HashBlock(b)
	pubKey := crypto.PublicKeyFromBytes(b.PublicKey)
	sigBytes := b.Signature
	sig := crypto.SignatureFromBytes(sigBytes)
	validSignarure := sig.Verify(pubKey, hash)
	if !validSignarure {
		return false, fmt.Errorf("invalid signature")
	}
	return true, nil
}

func SignBlock(pk *crypto.PrivateKey, block *proto.Block) *crypto.Signature {
	if len(block.Transactions) > 0 {
		tree, err := GetMerkleTree(block)
		if err != nil {
			panic(err)
		}
		block.Header.RootHash = GetRootHash(*tree)
	}
	hash := HashBlock(block)
	sig := pk.Sign(hash)
	block.PublicKey = pk.Public().Bytes()
	block.Signature = sig.Bytes()

	return sig
}

func VerifyRootHash(b *proto.Block) (bool, error) {
	tree, err := GetMerkleTree(b)
	if err != nil {
		return false, err
	}
	valid, err := tree.VerifyTree()
	if err != nil {
		return false, err
	}

	if !valid {
		return false, fmt.Errorf("merkele tree is not valie")
	}

	eq := bytes.Equal(b.Header.RootHash, GetRootHash(*tree))
	if !eq {
		return false, fmt.Errorf("block root hash %s and merkle root hash %s are not equal", hex.EncodeToString(b.Header.RootHash), hex.EncodeToString(tree.MerkleRoot()))
	}
	return eq, nil
}

func GetMerkleTree(b *proto.Block) (*merkletree.MerkleTree, error) {
	list := make([]merkletree.Content, len(b.Transactions))
	for i := 0; i < len(b.Transactions); i++ {
		list[i] = NewTxHash(HashTransaction(b.Transactions[i]))
	}

	tree, err := merkletree.NewTree(list)
	if err != nil {
		return nil, err
	}
	return tree, nil
}

func GetRootHash(tree merkletree.MerkleTree) []byte {
	rootHash := tree.MerkleRoot()
	return rootHash
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
