package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Phyrexian Broodlings — Creature — Phyrexian Minion {1}{B}{B}, 2/2
// (Jumpstart, #4892):
//
//	"{1}, Sacrifice a creature: Put a +1/+1 counter on this creature."
//
// Bloodflow Connoisseur's outlet with a {1} price. The Broodlings may
// sacrifice themselves: legal, and the counter then has nowhere to go.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7d5e4033-bd58-4854-b9d3-a440240995ee",
		Name:         "Phyrexian Broodlings",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}, Sacrifice a creature: Put a +1/+1 counter on this creature.",
			Purpose: game.Purpose{Answers: game.AnswerPump | game.AnswerSacOutlet},
			Cost:    Plus(ManaCost("{1}"), SacrificeACreature()),
			Effect:  Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1}),
		}},
	})
}
