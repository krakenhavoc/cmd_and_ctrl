package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Final Flourish — Instant {1}{B} (March of the Machine, S82 #5477):
//
//	"Kicker—Sacrifice an artifact or creature. (You may sacrifice an
//	 artifact or creature in addition to any other costs as you cast this
//	 spell.)
//	 Target creature gets -2/-2 until end of turn. If this spell was kicked,
//	 that creature gets -6/-6 until end of turn instead."
//
// The sacrifice is the kicker (CR 702.33), paid on casting; kicked, the
// creature gets -6/-6 INSTEAD of -2/-2, never both.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "63a85361-56bc-4209-aea8-b49319bc7b92",
		Name:          "Final Flourish",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.KickerSacrifice("an artifact or creature", effects.Or(effects.Artifact(), effects.Creature()))},
		Targets:       effects.TargetCreature("target creature"),
		Purpose:       effects.ForTargets(effects.HarmUntilEOT(0, -2, -2)),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if ctx.WasKicked() {
				return effects.VsPumpTheTarget(ctx, -6, -6, "Final Flourish — kicked: -6/-6")
			}
			return effects.VsPumpTheTarget(ctx, -2, -2, "Final Flourish — -2/-2")
		},
	})
}
