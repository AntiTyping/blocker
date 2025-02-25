package types

import (
	"blocker/util"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddBlock(t *testing.T) {
	chain := NewChain(NewMemoryBlockStore())
	block := util.RandomBlock()
	blockHash := HashBlock(block)

	assert.Nil(t, chain.AddBlock(block))
	actualBlock, err := chain.GetBlockByHash(blockHash)
	assert.Nil(t, err)
	assert.Equal(t, block, actualBlock)
}

func TestHeight(t *testing.T) {
	chain := NewChain(NewMemoryBlockStore())

	assert.Equal(t, 1, chain.Height())

	for i := 0; i < 99; i++ {
		chain.AddBlock(util.RandomBlock())
	}

	assert.Equal(t, 100, chain.Height())
}

func TestGetBlockByHash(t *testing.T) {
	chain := NewChain(NewMemoryBlockStore())

	first := util.RandomBlock()
	chain.AddBlock(first)
	actualFirst, err := chain.GetBlockByHash(HashBlock(first))
	assert.Nil(t, err)
	assert.Equal(t, first, actualFirst)

	second := util.RandomBlock()
	chain.AddBlock(second)
	actualSecond, err := chain.GetBlockByHash(HashBlock(second))
	assert.Nil(t, err)
	assert.Equal(t, second, actualSecond)
}

func TestGetBlockByHeight(t *testing.T) {
	chain := NewChain(NewMemoryBlockStore())

	first := util.RandomBlock()
	chain.AddBlock(first)
	actualFirst, err := chain.GetBlockByHeight(1)
	assert.Nil(t, err)
	assert.Equal(t, first, actualFirst)

	second := util.RandomBlock()
	chain.AddBlock(second)
	actualSecond, err := chain.GetBlockByHeight(2)
	assert.Nil(t, err)
	assert.Equal(t, second, actualSecond)
}
