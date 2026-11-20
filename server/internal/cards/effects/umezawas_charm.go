package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "7cfd7606-1dab-4627-ba42-37d6e375f9f1",
		Name:         "Umezawa's Charm",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeWithPurpose(ModeDoing("Target creature gets +2/+2 until end of turn.",
				TargetCreature("target creature"),
				BoostTheModesTarget(2, 2, "Umezawa's Charm — +2/+2")), ForTargets(BenefitUntilEOT(0, 2, 2))),
			ModeWithPurpose(ModeDoing("Target creature gets -1/-1 until end of turn.",
				TargetCreature("target creature"),
				BoostTheModesTarget(-1, -1, "Umezawa's Charm — -1/-1")), ForTargets(HarmUntilEOT(0, -1, -1))),
			ModeDoing("You gain 2 life.", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error { return GainLife{Amount: 2}.Apply(ctx) }),
		),
	})
}
