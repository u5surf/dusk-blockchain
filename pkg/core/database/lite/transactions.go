package lite

import (
	"encoding/hex"
	"gitlab.dusk.network/dusk-core/dusk-go/pkg/crypto/merkletree"
	"gitlab.dusk.network/dusk-core/dusk-go/pkg/p2p/wire/payload/block"
)

// Tx Transaction layer
type Tx struct {
	writable bool
	db       *DB
}

// GetBlockHeaderByHash gives the block header from the hash
func (t Tx) GetBlockHeaderByHash(hash []byte) (*block.Header, error) {

	// only a dummy select
	hex := hex.EncodeToString(hash)
	query := "SELECT height FROM HEADERS WHERE hash = " + hex
	rows, err := t.db.storage.Query(query)

	if err != nil {
		return nil, err
	}

	h := &block.Header{}
	for rows.Next() {
		err = rows.Scan(&h.Height)
	}

	return h, err
}

// WriteHeader writes dummy data
func (t Tx) WriteHeader(header *block.Header) error {

	//  only a dummy insert
	stmt, err := t.db.storage.Prepare("INSERT INTO HEADERS (hash, height ) values(?,?)")
	hex := hex.EncodeToString(header.Hash)

	if err == nil {
		_, err = stmt.Exec(hex, header.Height)
	}

	return err
}

func (t Tx) Commit() error {
	// Sqlite provides ACID support already
	return nil
}

func (t Tx) Rollback() error {
	// Sqlite provides ACID support already
	return nil
}

func (t Tx) Close() {
}

func (t Tx) FetchBlockHeader(hash []byte) (*block.Header, error) {
	return nil, nil
}
func (t Tx) FetchBlockTransactions(hash []byte) ([]merkletree.Payload, error) {
	return nil, nil
}

// FetchBlockHeaderByRound not implemented
func (t Tx) FetchBlockExists(header *block.Header) (bool, error) {
	return false, nil
}

// StoreBlock not implemented
func (t Tx) StoreBlock(block *block.Block) error {
	return nil
}
