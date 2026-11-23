package effectstest

import (
	"sync"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- the ordering probe ---------------------------------------------

const FirstLossProbeOracle = "test-first-life-loss-probe"

// firstLossSample is what a trigger's AppliesTo saw for one life-loss
// event: the loss itself and the controller's tally at that moment.
type FirstLossSample struct {
	Kind  game.EventKind
	Lost  int
	Tally int
}

var (
	FirstLossProbeMu      sync.Mutex
	FirstLossProbeSamples []FirstLossSample
)

func init() {
	effects.Register(effects.Spec{
		OracleID: FirstLossProbeOracle,
		Name:     "First Life Loss Probe",
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventChangeLife, game.EventDealDamage},
			Key:     "First Life Loss Probe",
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if lost, ok := effects.S22PlayerLostLife(ev, source.Controller, g); ok {
					FirstLossProbeMu.Lock()
					FirstLossProbeSamples = append(FirstLossProbeSamples, FirstLossSample{
						Kind: ev.Kind, Lost: lost, Tally: g.TurnTallyFor(source.Controller).LifeLost,
					})
					FirstLossProbeMu.Unlock()
				}
				return false
			},
			Effect: func(*game.Game, *game.StackItem) error { return nil },
		}},
	})
}
