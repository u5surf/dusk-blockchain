package consensus

import (
	"encoding/binary"
	"errors"
	"time"

	"gitlab.dusk.network/dusk-core/dusk-go/pkg/p2p/wire/payload/block"
	"gitlab.dusk.network/dusk-core/dusk-go/pkg/p2p/wire/payload/consensusmsg"

	"gitlab.dusk.network/dusk-core/dusk-go/pkg/core/consensus/zkproof"
	"gitlab.dusk.network/dusk-core/dusk-go/pkg/crypto/hash"
	"gitlab.dusk.network/dusk-core/dusk-go/pkg/p2p/wire/payload"
)

// Refactor of code made by jules

// GenerateBlock will generate a blockMsg and ScoreMsg
// if node is eligible.
func GenerateBlock(ctx *Context) error {

	err := generateParams(ctx)
	if err != nil {
		return err
	}

	// check threshold (eligibility)
	if ctx.Q <= ctx.Tau {
		return errors.New("Score is less than tau (threshold)")
	}

	// Generate ZkProof and Serialise
	// XXX: Prove may return error, so chaining like this may not be possible once zk implemented
	zkBytes, err := zkproof.Prove(ctx.X, ctx.Y, ctx.Z, ctx.M, ctx.k, ctx.Q, ctx.d).Bytes()
	if err != nil {
		return err
	}

	// Seed is the candidate signature of the previous seed
	seed, err := ctx.BLSSign(ctx.Keys.BLSSecretKey, ctx.Keys.BLSPubKey, ctx.LastHeader.Seed)
	if err != nil {
		return err
	}

	ctx.Seed = seed

	// Generate candidate block
	candidateBlock, err := newCandidateBlock(ctx)
	if err != nil {
		return err
	}

	// Create score msg
	pl, err := consensusmsg.NewCandidateScore(
		ctx.Q,                      // score
		zkBytes,                    // zkproof
		candidateBlock.Header.Hash, // candidateHash
		ctx.Seed,                   // seed for this round // XXX(TOG): could we use round number/Block height?
	)
	if err != nil {
		return err
	}

	sigEd, err := createSignature(ctx, pl)
	if err != nil {
		return err
	}

	msgScore, err := payload.NewMsgConsensus(ctx.Version, ctx.Round, ctx.LastHeader.Hash,
		ctx.Step, sigEd, []byte(*ctx.Keys.EdPubKey), pl)
	if err != nil {
		return err
	}

	if err := ctx.SendMessage(ctx.Magic, msgScore); err != nil {
		return err
	}

	// Create candidate msg
	pl2 := consensusmsg.NewCandidate(candidateBlock)
	msgCandidate, err := payload.NewMsgConsensus(ctx.Version, ctx.Round, ctx.LastHeader.Hash,
		ctx.Step, sigEd, []byte(*ctx.Keys.EdPubKey), pl2)
	if err != nil {
		return err
	}

	if err := ctx.SendMessage(ctx.Magic, msgCandidate); err != nil {
		return err
	}

	return nil
}

// generate M, X, Y, Z, Q
func generateParams(ctx *Context) error {
	// XXX: generating X, Y, Z in this way is in-efficient. Passing the parameters in directly from previous computation is better.
	// Wait until, specs for this has been semi-finalised
	M, err := generateM(ctx.k)
	if err != nil {
		return err
	}
	X, err := generateX(ctx.d, ctx.k)
	if err != nil {
		return err
	}
	Y, err := generateY(ctx.d, ctx.LastHeader.Seed, ctx.k)
	if err != nil {
		return err
	}
	Z, err := generateZ(ctx.LastHeader.Seed, ctx.k)
	if err != nil {
		return err
	}
	Q, err := GenerateScore(ctx.d, Y)
	if err != nil {
		return err
	}

	ctx.M = M
	ctx.X = X
	ctx.Y = Y
	ctx.Z = Z
	ctx.Q = Q

	return nil

}

// M = H(k)
func generateM(k []byte) ([]byte, error) {
	M, err := hash.Sha3256(k)
	if err != nil {
		return nil, err
	}

	return M, nil
}

// X = H(d, M)
func generateX(d uint64, k []byte) ([]byte, error) {

	M, err := generateM(k)

	dM := make([]byte, 8, 40)

	binary.LittleEndian.PutUint64(dM, d)

	dM = append(dM, M...)

	X, err := hash.Sha3256(dM)
	if err != nil {
		return nil, err
	}

	return X, nil
}

// Y = H(S, X)
func generateY(d uint64, S, k []byte) ([]byte, error) {

	X, err := generateX(d, k)
	if err != nil {
		return nil, err
	}

	SX := make([]byte, 0, 64) // X = 32 , prevSeed = 32
	SX = append(SX, S...)
	SX = append(SX, X...)

	Y, err := hash.Sha3256(SX)
	if err != nil {
		return nil, err
	}

	return Y, nil
}

// Z = H(S, M)
func generateZ(S, k []byte) ([]byte, error) {

	M, err := generateM(k)

	SM := make([]byte, 0, 64) // M = 32 , prevSeed = 32
	SM = append(SM, S...)
	SM = append(SM, M...)

	Z, err := hash.Sha3256(SM)
	if err != nil {
		return nil, err
	}

	return Z, nil
}

func newCandidateBlock(ctx *Context) (*block.Block, error) {

	candidateBlock := block.NewBlock()

	candidateBlock.SetPrevBlock(ctx.LastHeader)

	candidateBlock.Header.Seed = ctx.Seed
	candidateBlock.Header.Height = ctx.Round
	candidateBlock.Header.Timestamp = time.Now().Unix()

	// XXX: Generate coinbase/reward beforehand
	// Coinbase is still not decided
	txs := ctx.GetAllTXs()
	for _, tx := range txs {
		candidateBlock.AddTx(tx)
	}
	err := candidateBlock.SetRoot()
	if err != nil {
		return nil, err
	}

	err = candidateBlock.SetHash()
	if err != nil {
		return nil, err
	}

	return candidateBlock, nil
}
