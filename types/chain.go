package types

import (
	"blocker/crypto"
	"blocker/proto"
	"bytes"
	"encoding/hex"
	"fmt"
)

const goldenSeed = "183d81f40dd7d9233696dfa5e6eb8a287b1370f236efe844b3ed6c8d4896f6ce"
const genesisSupply = 1000

type HeaderList struct {
	headers []*proto.Header
}

func NewHeaderList() *HeaderList {
	return &HeaderList{
		headers: []*proto.Header{},
	}
}

func (l *HeaderList) Add(h *proto.Header) {
	l.headers = append(l.headers, h)
}

func (l *HeaderList) GetByHeight(height int) (*proto.Header, error) {
	if height < 0 && height > l.Height() {
		return nil, fmt.Errorf("no block found at height %d", height)
	}
	header := l.headers[height]

	return header, nil
}

func (l *HeaderList) Len() int {
	return len(l.headers)
}

func (l *HeaderList) Height() int {
	return l.Len()
}

type UTXO struct {
	Hash     string
	OutIndex int
	Amount   int64
	Spent    bool
}

type Chain struct {
	blockStore BlockStorer
	txStore    TXStorer
	uxtoStore  UTXOStorer
	headers    *HeaderList
}

func NewChain(bs BlockStorer, ts TXStorer) *Chain {
	chain := &Chain{
		blockStore: bs,
		txStore:    ts,
		uxtoStore:  NewMemoryUTXOStore(),
		headers:    NewHeaderList(),
	}
	chain.addBlock(createGenesisBlock())
	return chain
}

func (c *Chain) AddBlock(b *proto.Block) error {
	if err := c.ValidateBlock(b); err != nil {
		return err
	}
	return c.addBlock(b)
}

func (c *Chain) addBlock(b *proto.Block) error {
	c.headers.Add(b.Header)

	for _, tx := range b.Transactions {
		err := c.txStore.Put(tx)
		if err != nil {
			return err
		}

		hash := hex.EncodeToString(HashTransaction(tx))
		for idx, output := range tx.Outputs {
			utxo := &UTXO{
				Hash:     hash,
				OutIndex: idx,
				Amount:   output.Amount,
				Spent:    false,
			}
			c.uxtoStore.Put(utxo)
		}
	}
	return c.blockStore.Put(b)
}

func (c *Chain) Height() int {
	return c.headers.Height() - 1
}

func (c *Chain) GetBlockByHeight(height int) (*proto.Block, error) {
	header, err := c.headers.GetByHeight(height)
	if err != nil {
		panic(err)
	}
	block, err := c.GetBlockByHash(HashHeader(header))
	if err != nil {
		panic(err)
	}

	return block, nil
}

func (c *Chain) GetBlockByHash(hash []byte) (*proto.Block, error) {
	hasHex := hex.EncodeToString(hash)
	return c.blockStore.Get(hasHex)
}

func (c *Chain) ValidateBlock(b *proto.Block) error {
	// validate the signature of the block
	if !VerifyBlock(b) {
		return fmt.Errorf("invalid block")
	}

	// validate if the prevHas is the actual has of the current block
	currentBlock, err := c.GetBlockByHeight(c.Height())
	if err != nil {
		return err
	}
	hash := HashBlock(currentBlock)
	if !bytes.Equal(hash, b.Header.PreviousHash) {
		return fmt.Errorf("invlid previous hash")
	}

	// validate transactions
	for _, tx := range b.Transactions {
		if err := c.ValidateTransaction(tx); err != nil {

		}
	}
	return nil
}

func (c *Chain) ValidateTransaction(tx *proto.Transaction) error {
	if !VerifyTransaction(tx) {
		return fmt.Errorf("invalid tx ")
	}

	txHash := hex.EncodeToString(HashTransaction(tx))
	// check if all outputs are present and unspent
	nOutputs := len(tx.Outputs)
	var sumOutputs int64
	for i := 0; i < nOutputs; i++ {
		key := fmt.Sprintf("%s_%d", txHash, i)
		utxo, err := c.uxtoStore.Get(key)
		if err != nil {
			return err
		}
		if utxo.Spent {
			return fmt.Errorf("output is already spent")
		}
		sumOutputs += tx.Outputs[i].Amount
	}

	// check if all inputs are unspent
	nInputs := len(tx.Inputs)
	var sumInputs int64
	for i := 0; i < nInputs; i++ {
		key := fmt.Sprintf("%s_%d", hex.EncodeToString(tx.Inputs[i].PrevTxHash), tx.Inputs[i].PrevOutIndex)
		utxo, err := c.uxtoStore.Get(key)
		if err != nil {
			return err
		}
		if utxo.Spent {
			return fmt.Errorf("input is already delayed")
		}
		sumInputs += utxo.Amount
	}
	if sumInputs != sumOutputs {
		return fmt.Errorf("insufficient balance inputs are %d and outputs are %d", sumInputs, sumOutputs)
	}
	return nil
}

func createGenesisBlock() *proto.Block {
	privKey := crypto.NewPrivateKeyFromString(goldenSeed)

	tx := proto.Transaction{
		Version: 1,
		Inputs:  []*proto.TxInput{},
		Outputs: []*proto.TxOutput{
			{
				Amount:    1000,
				ToAddress: privKey.Public().Address().Bytes(),
			},
		},
	}
	block := &proto.Block{
		Header: &proto.Header{
			Version: 1,
		},
		Transactions: []*proto.Transaction{&tx},
	}
	SignBlock(privKey, block)
	return block
}
