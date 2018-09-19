package wallet

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCorrectVerif(t *testing.T) {
	priv, _ := NewPrivateKey()

	msg := []byte("hello world")

	sig, err := priv.Sign(msg)

	res := priv.Public().Verify(msg, sig)

	assert.Equal(t, nil, err)
	assert.Equal(t, true, res)
}
func TestWrongMessage(t *testing.T) {
	priv, _ := NewPrivateKey()

	msg := []byte("hello world")

	sig, _ := priv.Sign(msg)

	res := priv.Public().Verify([]byte("hello"), sig)
	assert.Equal(t, false, res)
}
func TestPubKeyToAddress(t *testing.T) {
	key := []byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}

	addr, err := pubKeyToAddress(key)
	assert.Equal(t, nil, err)
	assert.Equal(t, "DUSK1ABQgdjQpBGzXWQM4Cy7c7wMD3TK54Ec4mCieda5X79cJQuSLwo", addr)
}
