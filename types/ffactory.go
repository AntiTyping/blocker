package types

import (
	"blocker/crypto"
	"blocker/proto"
	crand "crypto/rand"
	"io"
)

type Factory struct {
}

func (f Factory) CreatePrivateKey() *crypto.PrivateKey {
	return crypto.GeneratePrivateKey()
}

func (f Factory) CreateGenesisPrivateKey() *crypto.PrivateKey {
	return crypto.NewPrivateKeyFromString(goldenSeed)
}

func (f Factory) CreatePublicKey() *crypto.PublicKey {
	return f.CreatePrivateKey().Public()
}

func (f Factory) CreateAddress() []byte {
	return f.CreatePublicKey().Address().Bytes()
}

func (f Factory) CreateHash() []byte {
	hash := [32]byte{}
	io.ReadFull(crand.Reader, hash[:])
	return hash[:]
}

func (f Factory) CreateTransaction() *proto.Transaction {
	return f.CreateTransactionWithAmount(1)
}

func (f Factory) CreateTransactionWithAmount(amount int64) *proto.Transaction {
	tx := proto.Transaction{
		Version: 1,
		Inputs: []*proto.TxInput{
			{
				PrevTxHash:   f.CreateHash(),
				PrevOutIndex: 0,
				PublicKey:    f.CreatePublicKey().Bytes(),
				Signature:    nil,
			},
		},
		Outputs: []*proto.TxOutput{
			{
				Amount:    amount,
				ToAddress: f.CreateAddress(),
			},
		},
	}
	return &tx
}
