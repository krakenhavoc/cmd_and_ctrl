package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ticked Off — {1}{R} Sorcery (Teenage Mutant Ninja Turtles Commander,
// #5011):
//
//	"Target creature gains double strike until end of turn."
//
// A layer-6 grant pinned to the target until the cleanup step (CR 514.2).
// A target that left or stopped being a creature gets nothing, and with no
// legal target the spell does not resolve (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "91b89c21-5f35-4aa5-925b-c756ec6969b4",
		Name:         "Ticked Off",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "double strike")),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.GrantEachLegalTargetUntilEOT(ctx, "double strike", "Ticked Off — double strike until end of turn")
		},
	})
}
