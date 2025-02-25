package types

import (
	"blocker/crypto"
	"blocker/proto"
	"blocker/util"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashTransaction(t *testing.T) {
	toPrivKey := crypto.GeneratePrivateKey()
	fromPrivKey := crypto.GeneratePrivateKey()
	fromPubKey := fromPrivKey.Public()
	input := &proto.TxInput{
		PrevTxHash:   util.RandomHash(),
		PrevOutIndex: 0,
		PublicKey:    fromPubKey.Bytes(),
	}
	output1 := &proto.TxOutput{
		Amount:    5,
		ToAddress: toPrivKey.Public().Address().Bytes(),
	}
	output2 := &proto.TxOutput{
		Amount:    95,
		ToAddress: fromPrivKey.Public().Address().Bytes(),
	}
	tx := &proto.Transaction{
		Version: 1,
		Inputs:  []*proto.TxInput{input},
		Outputs: []*proto.TxOutput{output1, output2},
	}

	sig := SignTransaction(fromPrivKey, tx)
	input.Signature = sig.Bytes()

	assert.Equal(t, 64, len(sig.Bytes()))
	assert.True(t, VerifyTransaction(tx))
}
