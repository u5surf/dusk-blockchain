package rangeproof

import (
	"math/big"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/toghrulmaharramov/dusk-go/ristretto"
)

func TestProveBulletProof(t *testing.T) {

	n := 20

	for i := 0; i < n; i++ {

		var amount ristretto.Scalar

		n := rand.Int63()
		amount.SetBigInt(big.NewInt(n))

		// Prove
		p, err := Prove(amount)

		assert.Equal(t, nil, err)

		// Verify
		ok, err := Verify(p)
		assert.Equal(t, nil, err)
		assert.Equal(t, true, ok)
	}

}

func TestComputeTau(t *testing.T) {
	a := ristretto.Scalar{}
	a.SetBigInt(big.NewInt(1))
	b := ristretto.Scalar{}
	b.SetBigInt(big.NewInt(2))
	c := ristretto.Scalar{}
	c.SetBigInt(big.NewInt(1))
	d := ristretto.Scalar{}
	d.SetBigInt(big.NewInt(1))
	e := ristretto.Scalar{}
	e.SetBigInt(big.NewInt(1))

	res := computeTaux(a, b, c, d, e)

	assert.Equal(t, int64(6), res.BigInt().Int64())

}

func TestComputeMu(t *testing.T) {
	var one ristretto.Scalar
	one.SetOne()

	var expected ristretto.Scalar
	expected.SetBigInt(big.NewInt(2))

	res := computeMu(one, one, one)

	ok := expected.Equals(&res)

	assert.Equal(t, true, ok)
}

/*
TODO: test values over the N threshold and named errors named errors
input: 2^N+1
output: error: value too large
*/
