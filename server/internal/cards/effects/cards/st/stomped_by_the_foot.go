package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stomped by the Foot — {1}{B} Instant (Teenage Mutant Ninja Turtles,
// #4457):
//
//	"Kicker—Sacrifice an artifact or creature. (You may sacrifice an
//	 artifact or creature in addition to any other costs as you cast this
//	 spell.)
//	 Target creature gets -2/-2 until end of turn. If this spell was kicked,
//	 that creature gets -5/-5 until end of turn instead."
//
// A non-mana kicker (CR 702.33a; Gatekeeper of Malakir's KickerSacrifice),
// paid as the spell is cast, read back with ctx.WasKicked() as it resolves.
// Kicked, the creature gets -5/-5 instead of -2/-2, never both.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "10185447-a13f-40d8-a8b6-39ce801f2fc0",
		Name:          "Stomped by the Foot",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.KickerSacrifice("an artifact or creature", effects.Or(effects.Artifact(), effects.Creature()))},
		Targets:       effects.TargetCreature("target creature"),
		Purpose:       effects.ForTargets(effects.HarmUntilEOT(0, -2, -2)),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if ctx.WasKicked() {
				return effects.VsPumpTheTarget(ctx, -5, -5, "Stomped by the Foot — kicked: -5/-5")
			}
			return effects.VsPumpTheTarget(ctx, -2, -2, "Stomped by the Foot — -2/-2")
		},
	})
}
