package types

import (
	"blocker/crypto"
	"blocker/proto"
	"blocker/util"
	crand "crypto/rand"
	"io"
	"math/rand"
	"time"
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

func (f Factory) CreateBlock() *proto.Block {
	chain := NewChain(NewMemoryBlockStore(), NewMemoryTXStore())
	privKey := Factory{}.CreateGenesisPrivateKey()
	to := Factory{}.CreateAddress()
	header := &proto.Header{
		Version:      1,
		Height:       int32(rand.Intn(1000)),
		PreviousHash: util.RandomHash(),
		RootHash:     util.RandomHash(),
		Timestamp:    time.Now().UnixNano(),
	}

	ftt, err := chain.txStore.Get("1c420f88a4b9f2c6c9615abedc6c1be07623b0ec71cde5e50be244250ccd5808")
	if err != nil {
		panic(err)
	}

	inputs := []*proto.TxInput{
		{
			PrevTxHash:   HashTransaction(ftt),
			PublicKey:    privKey.Public().Bytes(),
			PrevOutIndex: 0,
		},
	}
	outputs := []*proto.TxOutput{
		{
			Amount:    1000,
			ToAddress: to,
		},
		{
			Amount:    900,
			ToAddress: privKey.Public().Address().Bytes(),
		},
	}

	tx := proto.Transaction{
		Version: 1,
		Inputs:  inputs,
		Outputs: outputs,
	}

	inputs[0].Signature = privKey.Sign(HashTransaction(&tx)).Bytes()

	block := proto.Block{
		Header: header,
	}

	block.Transactions = []*proto.Transaction{&tx}
	return &block
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
