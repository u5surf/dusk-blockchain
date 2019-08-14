package generation_test

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/bwesterb/go-ristretto"
	"github.com/dusk-network/dusk-blockchain/pkg/core/block"
	"github.com/dusk-network/dusk-blockchain/pkg/core/consensus/generation"
	"github.com/dusk-network/dusk-blockchain/pkg/core/consensus/selection"
	"github.com/dusk-network/dusk-blockchain/pkg/core/consensus/user"
	"github.com/dusk-network/dusk-blockchain/pkg/core/tests/helper"
	"github.com/dusk-network/dusk-crypto"
	"github.com/dusk-network/dusk-crypto/key"
	"github.com/dusk-network/dusk-blockchain/pkg/p2p/wire/topics"
	zkproof "github.com/dusk-network/dusk-zkproof"
	"github.com/stretchr/testify/assert"
)

// Test that all of the functionality around score/proof generation works as intended.
// Note that the proof generator is mocked here, so the actual validity of the data
// is not tested.
func TestScoreGeneration(t *testing.T) {
	d := ristretto.Scalar{}
	d.Rand()
	k := ristretto.Scalar{}
	k.Rand()

	keys, _ := user.NewRandKeys()
	publicKey := key.NewKeyPair(nil).PublicKey()
	eb, streamer := helper.CreateGossipStreamer()

	gen := &mockGenerator{t}

	// launch score component
	generation.Launch(eb, nil, d, k, gen, gen, keys, publicKey)

	// send a block to start generation
	blk := helper.RandomBlock(t, 0, 1)
	b := new(bytes.Buffer)
	if err := blk.Encode(b); err != nil {
		t.Fatal(err)
	}
	eb.Publish(string(topics.AcceptedBlock), b)

	buf, err := streamer.Read()
	if err != nil {
		t.Fatal(err)
	}

	sev := &selection.ScoreEvent{Certificate: block.EmptyCertificate()}
	if err := selection.UnmarshalScoreEvent(bytes.NewBuffer(buf), sev); err != nil {
		t.Fatal(err)
	}

	// round should be 1
	assert.Equal(t, uint64(1), sev.Round)
}

// mockGenerator is used to test the generation component with the absence
// of the rust blind bid process.
type mockGenerator struct {
	t *testing.T
}

func (m *mockGenerator) GenerateBlock(round uint64, seed, proof, score, prevBlockHash []byte) (*block.Block, error) {
	return helper.RandomBlock(m.t, round, 10), nil
}

func (m *mockGenerator) GenerateProof(seed []byte) zkproof.ZkProof {
	proof, _ := crypto.RandEntropy(100)
	// add a score that will always pass threshold
	score, _ := big.NewInt(0).SetString("BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB", 16)
	z, _ := crypto.RandEntropy(32)
	bL, _ := crypto.RandEntropy(32)
	return zkproof.ZkProof{
		Proof:         proof,
		Score:         score.Bytes(),
		Z:             z,
		BinaryBidList: bL,
	}
}

func (m *mockGenerator) UpdateBidList(bl user.Bid)      {}
func (m *mockGenerator) RemoveExpiredBids(round uint64) {}
