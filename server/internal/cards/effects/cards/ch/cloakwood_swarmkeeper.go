package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cloakwood Swarmkeeper — {G} Creature — Elf Ranger 1/1
//
//	"Gathered Swarm — Whenever one or more tokens you control enter, put a
//	 +1/+1 counter on this creature."
//
// One trigger per batch of tokens entering together (CR 603.2c). The Swarmkeeper
// itself is no token, and a token an opponent controls doesn't count.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "61f09fcc-11cd-4999-a43a-54488b19861d",
		Name:         "Cloakwood Swarmkeeper",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OncePerBatch(effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := effects.EnteredUnderYourControl(ev, source, g, false)
				return ok && c.IsToken()
			}, "Gathered Swarm — put a +1/+1 counter on this creature", effects.PutACounterOnThis(game.CounterPlusOne))),
		},
	})
}
