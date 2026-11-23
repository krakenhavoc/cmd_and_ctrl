package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Grixis Charm — Instant {U}{B}{R} (Shards of Alara, #5588):
//
//	"Choose one —
//	 • Return target permanent to its owner's hand.
//	 • Target creature gets -4/-4 until end of turn.
//	 • Creatures you control get +2/+0 until end of turn."
//
// A modal spell (CR 700.2). The shrink and the pump are layer-7c effects
// until end of turn (CR 613.4c); the pump's set is the creatures its
// controller controls as it resolves (CR 611.2c). A targeted bullet whose
// target has gone does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "387ae7be-1cb5-4ae8-bb63-edf4941945a9",
		Name:         "Grixis Charm",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeDoing("Return target permanent to its owner's hand.",
				effects.TargetPermanent("target permanent"),
				effects.BounceTheModesTarget),
			effects.ModeWithPurpose(effects.ModeDoing("Target creature gets -4/-4 until end of turn.",
				effects.TargetCreature("target creature"),
				effects.BoostTheModesTarget(-4, -4, "Grixis Charm — -4/-4")),
				effects.ForTargets(effects.HarmUntilEOT(0, -4, -4))),
			effects.ModeDoing("Creatures you control get +2/+0 until end of turn.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.BoostUntilEOT{Match: effects.And(effects.Creature(), effects.YouControl()), Power: 2, Label: "Grixis Charm — +2/+0"}.Apply(ctx)
				}),
		),
	})
}
