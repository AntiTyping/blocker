package types

import (
	"blocker/crypto"
	"blocker/proto"
	"blocker/util"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	chain := NewChain(NewMemoryBlockStore(), NewMemoryTXStore())

	assert.Equal(t, 0, chain.Height())

	b, err := chain.GetBlockByHeight(0)
	assert.Nil(t, err)
	assert.NotNil(t, b)

}

func TestAddBlock(t *testing.T) {
	chain := NewChain(NewMemoryBlockStore(), NewMemoryTXStore())
	block := randomBlock(chain)
	blockHash := HashBlock(block)

	assert.Nil(t, chain.AddBlock(block))
	actualBlock, err := chain.GetBlockByHash(blockHash)
	assert.Nil(t, err)
	assert.Equal(t, block, actualBlock)
}

func TestHeight(t *testing.T) {
	chain := NewChain(NewMemoryBlockStore(), NewMemoryTXStore())

	assert.Equal(t, 0, chain.Height())

	for i := 0; i < 100; i++ {
		b := randomBlock(chain)
		chain.AddBlock(b)
	}

	assert.Equal(t, 100, chain.Height())
}

func TestGetBlockByHash(t *testing.T) {
	chain := NewChain(NewMemoryBlockStore(), NewMemoryTXStore())

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
	chain := NewChain(NewMemoryBlockStore(), NewMemoryTXStore())

	first := randomBlock(chain)
	err := chain.AddBlock(first)
	assert.Nil(t, err)
	actualFirst, err := chain.GetBlockByHeight(1)
	assert.Nil(t, err)
	assert.Equal(t, first, actualFirst)

	second := randomBlock(chain)
	chain.AddBlock(second)
	actualSecond, err := chain.GetBlockByHeight(2)
	assert.Nil(t, err)
	assert.Equal(t, second, actualSecond)
}

func TestCreateGenesisBlock(t *testing.T) {
	block := createGenesisBlock()

	assert.Equal(t, 1, len(block.Transactions))

	tx := block.Transactions[0]
	assert.Equal(t, int64(1000), tx.Outputs[0].Amount)

	sig := crypto.SignatureFromBytes(block.Signature)
	assert.True(t, sig.Verify(Factory{}.CreateGenesisPrivateKey().Public(), HashBlock(block)))
}

func TestAddBlockWithTx(t *testing.T) {
	// tx must be signed
	// cannot double spend
	// cannot overspend
	// must have previous output

	chain := NewChain(NewMemoryBlockStore(), NewMemoryTXStore())
	block := randomBlock(chain)
	privKey := Factory{}.CreateGenesisPrivateKey()
	to := Factory{}.CreateAddress()

	require.Equal(t, 0, chain.Height())

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
			Amount:    100,
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

	block.Transactions = []*proto.Transaction{&tx}

	tree, err := GetMerkleTree(block)
	assert.Nil(t, err)
	block.Header.RootHash = tree.MerkleRoot()

	inputs[0].Signature = privKey.Sign(HashTransaction(&tx)).Bytes()

	SignBlock(privKey, block)
	err = chain.AddBlock(block)
	require.Nil(t, err)
	require.Equal(t, 1, chain.Height())

	txHash := hex.EncodeToString(HashTransaction(&tx))

	fetchedTx, err := chain.txStore.Get(txHash)
	assert.Nil(t, err)
	assert.NotNil(t, fetchedTx)
	assert.Equal(t, tx, *fetchedTx)

	nOutputs := len(tx.Outputs)
	for i := 0; i < nOutputs; i++ {
		key := fmt.Sprintf("%s_%d", txHash, i)
		utxo, err := chain.uxtoStore.Get(key)
		assert.Nil(t, err)
		assert.Equal(t, outputs[i].Amount, utxo.Amount)
		assert.False(t, utxo.Spent)
	}
}

func TestValidateTransaction(t *testing.T) {
	chain := NewChain(NewMemoryBlockStore(), NewMemoryTXStore())
	block := randomBlock(chain)
	privKey := Factory{}.CreateGenesisPrivateKey()
	to := Factory{}.CreateAddress()

	require.Equal(t, 0, chain.Height())

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
			Amount:    100,
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

	block.Transactions = []*proto.Transaction{&tx}

	SignBlock(privKey, block)

	err = chain.AddBlock(block)
	require.Nil(t, err)
	require.Equal(t, 1, chain.Height())

	assert.Nil(t, chain.ValidateTransaction(&tx))
}

func TestAddTransactionWithInsufficientInputs(t *testing.T) {
	chain := NewChain(NewMemoryBlockStore(), NewMemoryTXStore())
	block := randomBlock(chain)
	privKey := Factory{}.CreateGenesisPrivateKey()
	to := Factory{}.CreateAddress()

	require.Equal(t, 0, chain.Height())

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

	block.Transactions = []*proto.Transaction{&tx}

	SignBlock(privKey, block)
	err = chain.AddBlock(block)
	require.Nil(t, err)
	require.Equal(t, 1, chain.Height())

	err = chain.ValidateTransaction(&tx)
	assert.NotNil(t, err)
	assert.Equal(t, fmt.Errorf("insufficient balance inputs are 1000 and outputs are 1900"), err)
}
