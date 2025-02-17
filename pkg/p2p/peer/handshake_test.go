package peer_test

import (
	"net"
	"testing"
	"time"

	_ "github.com/dusk-network/dusk-blockchain/pkg/core/database/lite"
	"github.com/dusk-network/dusk-blockchain/pkg/core/tests/helper"
	"github.com/dusk-network/dusk-blockchain/pkg/p2p/peer"
	"github.com/dusk-network/dusk-blockchain/pkg/p2p/peer/processing/chainsync"
	"github.com/dusk-network/dusk-blockchain/pkg/p2p/wire"
	"github.com/dusk-network/dusk-blockchain/pkg/p2p/wire/protocol"
)

func TestHandshake(t *testing.T) {

	eb := wire.NewEventBus()
	rpcBus := wire.NewRPCBus()
	counter := chainsync.NewCounter(eb)
	client, srv := net.Pipe()

	go func() {
		peerReader, err := helper.StartPeerReader(srv, eb, rpcBus, counter, nil)
		if err != nil {
			t.Fatal(err)
		}

		if err := peerReader.Accept(); err != nil {
			t.Fatal(err)
		}
	}()

	time.Sleep(500 * time.Millisecond)
	pw := peer.NewWriter(client, protocol.TestNet, eb)
	defer pw.Conn.Close()
	if err := pw.Handshake(); err != nil {
		t.Fatal(err)
	}
}
