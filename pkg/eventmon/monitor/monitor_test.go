package monitor_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net"
	"net/url"
	"os"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/stretchr/testify/assert"
	"gitlab.dusk.network/dusk-core/dusk-go/pkg/core/tests/helper"
	"gitlab.dusk.network/dusk-core/dusk-go/pkg/eventmon/logger"
	"gitlab.dusk.network/dusk-core/dusk-go/pkg/eventmon/monitor"
	"gitlab.dusk.network/dusk-core/dusk-go/pkg/p2p/wire"
	"gitlab.dusk.network/dusk-core/dusk-go/pkg/p2p/wire/topics"
)

const unixSoc = "unix:///tmp/dusk-socket"

func TestLogger(t *testing.T) {
	msgChan, addr := initTest()
	eb := wire.NewEventBus()
	conn, err := net.Dial("unix", addr)
	assert.NoError(t, err)

	testBlk := helper.RandomBlock(t, 23, 4)
	logProc := logger.New(eb, conn, nil)
	logProc.PublishBlockEvent(testBlk)

	result := <-msgChan

	assert.Equal(t, "monitor", result["process"])
	assert.Equal(t, float64(23), result["round"])

	_ = logProc.Close()
	// Make sure server is shut down
	<-msgChan
}

func TestSupervisor(t *testing.T) {
	msgChan, _ := initTest()
	eb := wire.NewEventBus()
	supervisor, err := monitor.Launch(eb, unixSoc)
	assert.NoError(t, err)

	testBuf := mockBlockBuf(t, 23)
	// testing that we can receive messages
	eb.Publish(string(topics.AcceptedBlock), testBuf)
	result := <-msgChan

	assert.Equal(t, "monitor", result["process"])
	assert.Equal(t, float64(23), result["round"])
	_ = supervisor.Stop()
	// Make sure server is shut down
	<-msgChan
}

func TestSupervisorReconnect(t *testing.T) {
	msgChan, addr := initTest()
	eb := wire.NewEventBus()
	supervisor, err := monitor.Launch(eb, unixSoc)
	assert.NoError(t, err)

	testBuf := mockBlockBuf(t, 23)
	// testing that we can receive messages
	eb.Publish(string(topics.AcceptedBlock), testBuf)
	<-msgChan

	assert.NoError(t, supervisor.Stop())
	// Make sure server is shut down
	<-msgChan

	testBuf = mockBlockBuf(t, 24)
	eb.Publish(string(topics.AcceptedBlock), testBuf)
	select {
	case <-msgChan:
		assert.FailNow(t, "Expected the supervised LogProcessor to be closed")
	case <-time.After(1 * time.Second):
		// all fine
	}

	// restarting the server as the Stop has likely killed it
	msgChan = startSrv(addr)
	// reconnecting the supervised process
	assert.NoError(t, supervisor.Reconnect())
	// messages streamed when the process is down are lost, so we need to send another message
	eb.Publish(string(topics.AcceptedBlock), testBuf)
	result := <-msgChan
	assert.Equal(t, "monitor", result["process"])
	assert.Equal(t, float64(24), result["round"])

	_ = supervisor.Stop()
	// Make sure server is shut down
	<-msgChan
}

func TestResumeRight(t *testing.T) {
	msgChan, _ := initTest()
	eb := wire.NewEventBus()
	supervisor, err := monitor.Launch(eb, unixSoc)
	assert.NoError(t, err)

	testBuf := mockBlockBuf(t, 23)
	eb.Publish(string(topics.AcceptedBlock), testBuf)
	round1 := <-msgChan
	if _, ok := round1["blockTime"]; ok {
		assert.FailNow(t, "First round should not really have a block time. Instead found %d", round1["blockTime"])
	}

	time.Sleep(time.Second)
	testBuf = mockBlockBuf(t, 24)
	eb.Publish(string(topics.AcceptedBlock), testBuf)
	round2 := <-msgChan

	assert.InDelta(t, float64(1), round2["blockTime"], float64(0.1))

	_ = supervisor.Stop()
	// Make sure server is shut down
	<-msgChan
}

func TestNotifyErrors(t *testing.T) {
	endChan := make(chan struct{})
	msgChan, _ := initTest()
	eb := wire.NewEventBus()
	supervisor, err := monitor.Launch(eb, unixSoc)
	assert.NoError(t, err)

	log.AddHook(supervisor)
	log.Errorln("pippo")

	// wrapped in a go routing to check that there are no race conditions
	go func() {
		msg := <-msgChan
		assert.Equal(t, "error", msg["level"])
		assert.Equal(t, "pippo", msg["msg"])
		endChan <- struct{}{}
	}()

	testBuf := mockBlockBuf(t, 23)
	eb.Publish(string(topics.AcceptedBlock), testBuf)
	result := <-msgChan
	assert.Equal(t, "monitor", result["process"])
	<-endChan
	_ = supervisor.Stop()
	// Make sure server is shut down
	<-msgChan
}

func mockBlockBuf(t *testing.T, height uint64) *bytes.Buffer {
	blk := helper.RandomBlock(t, height, 4)
	buf := new(bytes.Buffer)
	if err := blk.Encode(buf); err != nil {
		panic(err)
	}

	return buf
}

func initTest() (<-chan map[string]interface{}, string) {
	addr := unixSocPath()
	_ = os.Remove(addr)
	msgChan := startSrv(addr)
	return msgChan, addr
}

func startSrv(addr string) <-chan map[string]interface{} {
	msgChan := spinSrv(addr)
	// waiting for the server to be up and running
	<-msgChan
	return msgChan
}

func unixSocPath() string {
	uri, err := url.Parse(unixSoc)
	if err != nil {
		panic(err)
	}
	return uri.Path
}

func spinSrv(addr string) <-chan map[string]interface{} {
	resChan := make(chan map[string]interface{})

	go func() {
		var conn net.Conn
		srv, err := net.Listen("unix", addr)
		if err != nil {
			panic(err)
		}
		resChan <- nil

		conn, err = srv.Accept()
		// notifying that the server can accept connections
		if err != nil {
			panic(err)
		}

		if conn == nil {
			panic("Connection is nil")
		}

		// we create a decoder that reads directly from the socket
		d := json.NewDecoder(conn)
		for {
			var msg map[string]interface{}
			if err := d.Decode(&msg); err == io.EOF {
				break
			} else if err != nil {
				panic(err)
			}

			resChan <- msg
		}
		_ = conn.Close()
		srv.Close()
		resChan <- nil
	}()

	return resChan
}
