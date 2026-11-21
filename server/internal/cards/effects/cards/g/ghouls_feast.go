package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ghoul's Feast — Instant {1}{B} (Jumpstart 2022, #5948):
//
//	"Target creature gets +X/+0 until end of turn, where X is the number of
//	 creature cards in your graveyard."
//
// X is counted once as the spell resolves, in the caster's graveyard (the
// card itself is not a creature card and is still on the stack).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "042e0533-faf4-475e-be7b-438d23c6e605",
		Name:         "Ghoul's Feast",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0)),
		OnResolve:    effects.GapB170PlusXCreatureCards,
	})
}
