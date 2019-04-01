package main

import (
	"fmt"
	"os"
	"os/signal"

	"gitlab.dusk.network/dusk-core/dusk-go/pkg/core/block"
)

func main() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	srv := Setup()
	go srv.Listen()
	ips := ConnectToSeeder()
	connMgr := NewConnMgr(CmgrConfig{
		Port:     "8081",
		OnAccept: srv.OnAccept,
		OnConn:   srv.OnConnection,
	})

	// if we are the first, initialize consensus on round 1
	if len(ips) == 0 {
		fmt.Println("starting consensus")
		srv.StartConsensus(1)
	} else {
		for _, ip := range ips {
			if err := connMgr.Connect(ip); err != nil {
				fmt.Println(err)
			}

		}

		// get highest block
		var highest *block.Block
		for _, block := range srv.Blocks {
			if block.Header.Height > highest.Header.Height {
				highest = &block
			}
		}

		// if height is not 0, init consensus on 2 rounds after it
		// +1 because the round is always height + 1
		// +1 because we dont want to get stuck on a round thats currently happening
		if highest != nil && highest.Header.Height != 0 {
			srv.StartConsensus(highest.Header.Height + 2)
		} else {
			srv.StartConsensus(1)
		}
		fmt.Println("starting consensus")

	}

	// Wait until the interrupt signal is received from an OS signal or
	// shutdown is requested through one of the subsystems such as the RPC
	// server.
	<-interrupt
}
