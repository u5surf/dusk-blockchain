package generation

import (
	"bytes"
	"encoding/hex"

	"github.com/dusk-network/dusk-crypto"
	"github.com/dusk-network/dusk-crypto/key"
)

// GenerateGenesisBlock is a developer utility for regenerating the genesis block
// as they would be different per network type. Once a genesis block is
// approved, its hex blob should be copied into config.TestNetGenesisBlob
func GenerateGenesisBlock(generatorPubKey *key.PublicKey) (string, error) {
	g := &blockGenerator{
		rpcBus:    nil,
		genPubKey: generatorPubKey,
	}

	// TODO: do we need to generate correct proof and score
	seed, _ := crypto.RandEntropy(33)
	proof, _ := crypto.RandEntropy(32)
	score, _ := crypto.RandEntropy(32)

	b, err := g.GenerateBlock(0, seed, proof, score, make([]byte, 32))
	if err != nil {
		return "", err
	}

	buf := new(bytes.Buffer)
	if err := b.Encode(buf); err != nil {
		return "", err
	}

	return hex.EncodeToString(buf.Bytes()), nil
}
