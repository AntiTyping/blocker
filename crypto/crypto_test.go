package crypto

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGeneratePrivateKey(t *testing.T) {
	privKey := GeneratePrivateKey()

	assert.Equal(t, len(privKey.Bytes()), PrivKeyLen)

	pubKey := privKey.Public()

	assert.Equal(t, len(pubKey.Bytes()), PubKeyLen)
}

func TestNewPrivateKeyFromString(t *testing.T) {
	seed := "f270acaebd8a7153a38d61e8f64eb0f2695745d72501725ca4a9632818fbbd1e"
	privKey := NewPrivateKeyFromString(seed)

	assert.Equal(t, PrivKeyLen, len(privKey.Bytes()))

	assert.Equal(t, "b0a03971455014435c4c833e20053b4e281c65bb", privKey.Public().Address().String())
}

func TestPrivateKeySing(t *testing.T) {
	privKey := GeneratePrivateKey()
	msg := []byte("foo bar")

	sig := privKey.Sign(msg)

	assert.True(t, sig.Verify(privKey.Public(), msg))

	// test with invalid msg
	assert.False(t, sig.Verify(privKey.Public(), []byte("foo")))

	// test with invalid key
	assert.False(t, sig.Verify(GeneratePrivateKey().Public(), msg))
}

func TestPublicKeyToAddress(t *testing.T) {
	privKey := GeneratePrivateKey()
	pubKey := privKey.Public()
	address := pubKey.Address()

	assert.Equal(t, AddressLen, len(address.Bytes()))
	fmt.Println(address)
}
