package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Enduring Scalelord — Creature — Dragon {4}{G}{W}, 4/4:
//
//	"Flying
//	 Whenever one or more +1/+1 counters are put on another creature you
//	 control, you may put a +1/+1 counter on this creature."
//
// Wildwood Scourge's trigger with a "may": one EventCounterPlaced is one
// placement, so "one or more" fires once however many counters landed. The
// counters on the Scalelord itself never fire it ("another"), and a
// creature you do not control does not count. The Scalelord must still be on
// the battlefield when the trigger resolves; the choice is made then
// (CR 603.5).
//
// No simplification.
func init() {
	const label = "Enduring Scalelord — you may put a +1/+1 counter on this creature"
	trig := effects.On(game.EventCounterPlaced, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		return effects.PlusCountersOnAnotherCreatureYouControl(ev, source, g)
	}, label, effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1}))
	effects.Register(effects.Spec{
		OracleID:        "270a21c1-4c44-46ef-8c4f-e08484f360e1",
		Name:            "Enduring Scalelord",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Optional(trig, "Enduring Scalelord — put a +1/+1 counter on this creature?"),
		},
	})
}
