package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bloodthirsty Aerialist — Creature — Vampire Rogue {1}{B}{B}, 2/3
// (Foundations Jumpstart):
//
//	"Flying
//	 Whenever you gain life, put a +1/+1 counter on this creature."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ae78b6f9-8334-479e-8111-8e3acb2b7ca7",
		Name:            "Bloodthirsty Aerialist",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			LifeGainPaying(
				WheneverYouGainLife("Bloodthirsty Aerialist — put a +1/+1 counter on this creature", Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
				game.LifeGainPayoff{Counters: 1}),
		},
	})
}
