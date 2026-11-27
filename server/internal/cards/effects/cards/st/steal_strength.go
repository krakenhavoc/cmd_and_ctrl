package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Steal Strength — Instant {1}{B} (Prophecy, S82 OLD batch 674, #5617):
//
//	"Target creature gets +1/+1 until end of turn. Another target creature
//	 gets -1/-1 until end of turn."
//
// Leeching Bite's text with the colours swapped: two distinct targets, each
// half judged on its own target at resolution (CR 608.2b), so losing one does
// not stop the other.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e67434e6-5408-416a-be19-64291fca8e3e",
		Name:         "Steal Strength",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 1, 1), effects.HarmUntilEOT(1, -1, -1)),
		Targets: effects.Clauses(
			effects.TargetCreature("target creature"),
			effects.Distinct(effects.TargetCreature("another target creature")),
		),
		OnResolve: effects.S82OldB674PlusOnePlusOneAndMinusOneMinusOne("Steal Strength"),
	})
}
