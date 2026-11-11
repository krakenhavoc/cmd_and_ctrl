package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Grim Poppet — Artifact Creature — Scarecrow {7}, 4/4 (Lorwyn Eclipsed
// Commander, #5368):
//
//	"This creature enters with three -1/-1 counters on it.
//	 Remove a -1/-1 counter from this creature: Put a -1/-1 counter on
//	 another target creature."
//
// The counters are an entry replacement (b10EntersWithCounters), so it is
// a 1/1 the moment it arrives. Removing one is the cost, paid at announce
// (CR 118.3), and the target is any creature but the Poppet itself.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "43e0dbaf-3c81-47e2-b0c2-364f7ac7295f",
		Name:         "Grim Poppet",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			b10EntersWithCounters(game.CounterMinusOne, 3, "Grim Poppet: enters with three -1/-1 counters"),
		},
		Activated: []ActivatedAbility{{
			Label:   "Remove a -1/-1 counter from this creature: Put a -1/-1 counter on another target creature.",
			Cost:    RemoveCountersFromThis(game.CounterMinusOne, 1),
			Targets: Harmful(Another(TargetCreature("another target creature"))),
			Effect:  cmdr2B25CountersOnTarget(1),
		}},
	})
}
