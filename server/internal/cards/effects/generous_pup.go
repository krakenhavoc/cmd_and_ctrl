package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Generous Pup — Creature — Dog {1}{W}, 2/2:
//
//	"Vigilance
//	 Whenever one or more +1/+1 counters are put on this creature, put a
//	 +1/+1 counter on each other creature you control. This ability triggers
//	 only once each turn."
//
// Constable of the Realm's trigger (one EventCounterPlaced is one placement,
// whoever put the counters) held to once each turn per object, with Mikaeus's
// "each other creature you control". The set is fixed as the ability
// resolves; the counters the Pup puts on others never trigger it ("this").
//
// No simplification.
func init() {
	const label = "Generous Pup — put a +1/+1 counter on each other creature you control"
	Register(Spec{
		OracleID:        "9ee0d4a0-fd8d-43c4-a2aa-c4e6af5beecc",
		Name:            "Generous Pup",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			TriggersOnlyOnceEachTurn(On(game.EventCounterPlaced, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return plusCountersOnThis(ev, source, g)
			}, label, putCounterOnEachOtherCreatureYouControl)),
		},
	})
}
