package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vicious Offering — Instant {1}{B} (Dominaria, #5530):
//
//	"Kicker—Sacrifice a creature. (You may sacrifice a creature in addition to any other costs as you cast this spell.)
//	 Target creature gets -2/-2 until end of turn. If this spell was kicked, that creature gets -5/-5 until end of turn instead."
//
// Stomped by the Foot's shape with a creature as the sacrifice kicker:
// -2/-2, or -5/-5 when kicked.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "c6adfaa7-d9cf-4501-8e7d-e035369390d2",
		Name:          "Vicious Offering",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.KickerSacrifice("a creature", effects.Creature())},
		Targets:       effects.TargetCreature("target creature"),
		Purpose:       effects.ForTargets(effects.HarmUntilEOT(0, -2, -2)),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if ctx.WasKicked() {
				return effects.VsPumpTheTarget(ctx, -5, -5, "Vicious Offering — kicked: -5/-5")
			}
			return effects.VsPumpTheTarget(ctx, -2, -2, "Vicious Offering — -2/-2")
		},
	})
}
