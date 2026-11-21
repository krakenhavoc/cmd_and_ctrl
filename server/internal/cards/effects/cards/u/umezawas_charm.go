package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Umezawa's Charm — Instant {1}{B} (Modern Horizons, #5519):
//
//	"Choose one —
//	 • Target creature gets +2/+2 until end of turn.
//	 • Target creature gets -1/-1 until end of turn.
//	 • You gain 2 life."
//
// Life of Toshiro Umezawa's chapter, as an instant: three bullets on the
// shared modal builders.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7cfd7606-1dab-4627-ba42-37d6e375f9f1",
		Name:         "Umezawa's Charm",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeWithPurpose(effects.ModeDoing("Target creature gets +2/+2 until end of turn.",
				effects.TargetCreature("target creature"),
				effects.BoostTheModesTarget(2, 2, "Umezawa's Charm — +2/+2")), effects.ForTargets(effects.BenefitUntilEOT(0, 2, 2))),
			effects.ModeWithPurpose(effects.ModeDoing("Target creature gets -1/-1 until end of turn.",
				effects.TargetCreature("target creature"),
				effects.BoostTheModesTarget(-1, -1, "Umezawa's Charm — -1/-1")), effects.ForTargets(effects.HarmUntilEOT(0, -1, -1))),
			effects.ModeDoing("You gain 2 life.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.GainLife{Amount: 2}.Apply(ctx)
				}),
		),
	})
}
