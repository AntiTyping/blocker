package types

import (
	"blocker/proto"
	"encoding/hex"
	"fmt"
)

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
	return l.Len() + 1
}

type Chain struct {
	blockStore BlockStorer
	headers    *HeaderList
}

func NewChain(bs BlockStorer) *Chain {
	return &Chain{
		blockStore: bs,
		headers:    NewHeaderList(),
	}
}

func (c *Chain) AddBlock(b *proto.Block) error {
	// Validation
	c.headers.Add(b.Header)
	return c.blockStore.Put(b)
}

func (c *Chain) Height() int {
	return c.headers.Height()
}

func (c *Chain) GetBlockByHeight(height int) (*proto.Block, error) {
	header, err := c.headers.GetByHeight(height - 1)
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
