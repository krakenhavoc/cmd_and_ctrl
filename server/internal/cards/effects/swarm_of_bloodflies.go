package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Swarm of Bloodflies — Creature — Insect {4}{B}, 0/0 (Jumpstart):
//
//	"Flying
//	 This creature enters with two +1/+1 counters on it.
//	 Whenever another creature dies, put a +1/+1 counter on this
//	 creature."
//
// The counters are a self-replacement on entry (b10EntersWithCounters),
// on the creature before any state check sees a 0/0. The death trigger
// is Orca, Siege Demon's: any other creature, any controller, tokens
// included, read off the last-known battlefield characteristics.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "767e4aa9-527f-425f-bffb-6d842e1fbb74",
		Name:            "Swarm of Bloodflies",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Replacements: []game.ReplacementEffect{
			b10EntersWithCounters(game.CounterPlusOne, 2, "Swarm of Bloodflies: enters with two +1/+1 counters"),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b15AnotherCreatureDied(ev, source, g)
			}, "Swarm of Bloodflies — put a +1/+1 counter on it", Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
