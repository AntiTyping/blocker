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

type Chain struct {
	blockStore BlockStorer
	txStore    TXStorer
	headers    *HeaderList
}

func NewChain(bs BlockStorer, ts TXStorer) *Chain {
	chain := &Chain{
		blockStore: bs,
		txStore:    ts,
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
		fmt.Println(hex.EncodeToString(HashTransaction(tx)))
		if err != nil {
			return err
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
	if !VerifyBlock(b) {
		return fmt.Errorf("invalid block")
	}
	currentBlock, err := c.GetBlockByHeight(c.Height())
	if err != nil {
		return err
	}
	hash := HashBlock(currentBlock)
	if !bytes.Equal(hash, b.Header.PreviousHash) {
		return fmt.Errorf("invlid previous hash")
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
