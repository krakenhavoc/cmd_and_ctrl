package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Scion of the Swarm — Creature — Vampire Cleric {3}{B}{B}, 3/3
// (Foundations Jumpstart, S82 BNS):
//
//	"Flying
//	 Whenever you gain life, put a +1/+1 counter on this creature."
//
// Ajani's Pridemate's trigger: one counter per life-gain event, not per
// point of life.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ccaeabe1-42fa-4f29-bf80-d9b73498eec6",
		Name:            "Scion of the Swarm",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			LifeGainPaying(
				WheneverYouGainLife("Scion of the Swarm — put a +1/+1 counter on it", Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
				game.LifeGainPayoff{Counters: 1}),
		},
	})
}
