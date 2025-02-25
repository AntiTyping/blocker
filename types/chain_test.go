package types

import (
	"blocker/crypto"
	"blocker/proto"
	"blocker/util"
	"testing"

	"github.com/stretchr/testify/assert"
)

func randomBlock(chain *Chain) *proto.Block {
	block := util.RandomBlock()
	prevBlock, _ := chain.GetBlockByHeight(chain.Height())
	block.Header.PreviousHash = HashBlock(prevBlock)
	pk := crypto.GeneratePrivateKey()
	SignBlock(pk, block)

	return block
}

func TestNewChain(t *testing.T) {
	chain := NewChain(NewMemoryBlockStore())

	assert.Equal(t, 0, chain.Height())

	b, err := chain.GetBlockByHeight(0)
	assert.Nil(t, err)
	assert.NotNil(t, b)

}

func TestAddBlock(t *testing.T) {
	chain := NewChain(NewMemoryBlockStore())
	block := randomBlock(chain)
	blockHash := HashBlock(block)

	assert.Nil(t, chain.AddBlock(block))
	actualBlock, err := chain.GetBlockByHash(blockHash)
	assert.Nil(t, err)
	assert.Equal(t, block, actualBlock)
}

func TestHeight(t *testing.T) {
	chain := NewChain(NewMemoryBlockStore())

	assert.Equal(t, 0, chain.Height())

	for i := 0; i < 100; i++ {
		b := randomBlock(chain)
		chain.AddBlock(b)
	}

	assert.Equal(t, 100, chain.Height())
}

func TestGetBlockByHash(t *testing.T) {
	chain := NewChain(NewMemoryBlockStore())

	first := randomBlock(chain)
	chain.AddBlock(first)
	actualFirst, err := chain.GetBlockByHash(HashBlock(first))
	assert.Nil(t, err)
	assert.Equal(t, first, actualFirst)

	second := randomBlock(chain)
	chain.AddBlock(second)
	actualSecond, err := chain.GetBlockByHash(HashBlock(second))
	assert.Nil(t, err)
	assert.Equal(t, second, actualSecond)
}

func TestGetBlockByHeight(t *testing.T) {
	chain := NewChain(NewMemoryBlockStore())

	first := randomBlock(chain)
	chain.AddBlock(first)
	actualFirst, err := chain.GetBlockByHeight(1)
	assert.Nil(t, err)
	assert.Equal(t, first, actualFirst)

	second := randomBlock(chain)
	chain.AddBlock(second)
	actualSecond, err := chain.GetBlockByHeight(2)
	assert.Nil(t, err)
	assert.Equal(t, second, actualSecond)
}
