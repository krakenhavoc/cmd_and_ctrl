package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Stingmoggie — Creature — Elemental {3}{R}, 0/0 (Morningtide, #5592):
//
//	"This creature enters with two +1/+1 counters on it.
//	 {3}{R}, Remove a +1/+1 counter from this creature: Destroy target
//	 artifact or land."
//
// District Mascot's shape: the entry counters are a self-replacement
// (CR 614.1c), on the creature before any state check sees the printed
// 0/0, and the counter removal is part of the cost, paid at announce, so
// the ability cannot be activated without a counter (CR 602.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e868028f-4c6f-43d1-9cc6-6f3ce2df2ec3",
		Name:         "Stingmoggie",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			b10EntersWithCounters(game.CounterPlusOne, 2, "Stingmoggie: enters with two +1/+1 counters"),
		},
		Activated: []ActivatedAbility{{
			Label:   "{3}{R}, Remove a +1/+1 counter from this creature: Destroy target artifact or land.",
			Cost:    Plus(ManaCost("{3}{R}"), RemoveCountersFromThis(game.CounterPlusOne, 1)),
			Targets: TargetPermanent("target artifact or land", Or(Artifact(), Land())),
			Effect:  destroyFirstLegalTarget,
		}},
	})
}
