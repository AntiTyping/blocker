package types

import (
	"blocker/crypto"
	"blocker/util"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVerifyBlock(t *testing.T) {
	pk := crypto.GeneratePrivateKey()
	block := util.RandomBlock()
	block.Signature = SignBlock(pk, block).Bytes()

	verified, err := VerifyBlock(block)
	assert.Nil(t, err)
	assert.True(t, verified)

	invalidPk := crypto.GeneratePrivateKey()
	block.PublicKey = invalidPk.Public().Bytes()

	verified, err = VerifyBlock(block)
	assert.NotNil(t, err)
	assert.False(t, verified)
}

func TestHashBlock(t *testing.T) {
	block := util.RandomBlock()

	hash := HashBlock(block)

	assert.Equal(t, len(hash), 32)
}

func TestSignBlock(t *testing.T) {
	block := util.RandomBlock()
	privKey := crypto.GeneratePrivateKey()
	pubKey := privKey.Public()

	sig := SignBlock(privKey, block)

	assert.Equal(t, 64, len(sig.Bytes()))

	assert.True(t, sig.Verify(pubKey, HashBlock(block)))

	assert.Equal(t, pubKey.Bytes(), block.PublicKey)
}
