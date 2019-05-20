package reduction

import (
	"encoding/hex"
	"time"

	log "github.com/sirupsen/logrus"
	"gitlab.dusk.network/dusk-core/dusk-go/pkg/core/consensus"
	"gitlab.dusk.network/dusk-core/dusk-go/pkg/core/consensus/committee"
	"gitlab.dusk.network/dusk-core/dusk-go/pkg/core/consensus/selection"
	"gitlab.dusk.network/dusk-core/dusk-go/pkg/core/consensus/user"
	"gitlab.dusk.network/dusk-core/dusk-go/pkg/p2p/wire"
	"gitlab.dusk.network/dusk-core/dusk-go/pkg/p2p/wire/topics"
)

type (
	// Broker is the message broker for the reduction process.
	broker struct {
		filter      *consensus.EventFilter
		accumulator *consensus.Accumulator
		reducer     *reducer

		// utility context to group interfaces and channels to be passed around
		ctx *context

		// channels linked to subscribers
		roundUpdateChan <-chan uint64
		stepChan        <-chan struct{}
		selectionChan   <-chan *selection.ScoreEvent
	}
)

// LaunchReducer creates and wires a broker, initiating the components that
// have to do with Block Reduction
func LaunchReducer(eventBroker wire.EventBroker, committee committee.Committee, keys *user.Keys, timeout time.Duration) *broker {
	handler := newReductionHandler(committee)
	broker := newBroker(eventBroker, handler, committee, keys, timeout)

	go broker.Listen()
	return broker
}

func launchReductionFilter(eventBroker wire.EventBroker, ctx *context,
	accumulator *consensus.Accumulator) *consensus.EventFilter {

	filter := consensus.NewEventFilter(ctx.handler, ctx.state, accumulator, true)
	republisher := consensus.NewRepublisher(eventBroker, topics.Reduction)
	eventBroker.SubscribeCallback(string(topics.Reduction), filter.Collect)
	eventBroker.RegisterPreprocessor(string(topics.Reduction), republisher, &consensus.Validator{})
	return filter
}

// newBroker will return a reduction broker.
func newBroker(eventBroker wire.EventBroker, handler handler,
	committee committee.Committee, keys *user.Keys, timeout time.Duration) *broker {
	scoreChan := initBestScoreUpdate(eventBroker)
	ctx := newCtx(handler, committee, keys, timeout)
	accumulator := consensus.NewAccumulator(ctx.handler, consensus.NewAccumulatorStore())
	filter := launchReductionFilter(eventBroker, ctx, accumulator)
	roundChannel := consensus.InitRoundUpdate(eventBroker)
	stepSub := ctx.state.SubscribeStep()

	return &broker{
		roundUpdateChan: roundChannel,
		ctx:             ctx,
		filter:          filter,
		accumulator:     accumulator,
		selectionChan:   scoreChan,
		stepChan:        stepSub.StateChan,
		reducer:         newReducer(accumulator.CollectedVotesChan, ctx, eventBroker, accumulator),
	}
}

// Listen for incoming messages.
func (b *broker) Listen() {
	for {
		select {
		case round := <-b.roundUpdateChan:
			log.WithFields(log.Fields{
				"process": "reduction",
				"round":   round,
			}).Debug("Got round update")
			b.reducer.end()
			b.accumulator.Clear()
			b.filter.UpdateRound(round)
		case ev := <-b.selectionChan:
			if ev == nil {
				log.WithFields(log.Fields{
					"process": "reduction",
				}).Debug("got empty selection message")
				b.reducer.startReduction(make([]byte, 32))
				b.filter.FlushQueue()
			} else if ev.Round == b.ctx.state.Round() {
				log.WithFields(log.Fields{
					"process": "reduction",
					"hash":    hex.EncodeToString(ev.VoteHash),
				}).Debug("got selection message")
				b.reducer.startReduction(ev.VoteHash)
				b.filter.FlushQueue()
			} else {
				log.WithFields(log.Fields{
					"process":     "reduction",
					"event round": ev.Round,
				}).Debug("got obsolete selection message")
				b.reducer.startReduction(make([]byte, 32))
				b.filter.FlushQueue()
			}

		case <-b.stepChan:
			b.accumulator.Clear()
			b.filter.FlushQueue()
		}
	}
}
