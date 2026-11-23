package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dromar's Charm — Instant {W}{U}{B} (Planeshift, #5615):
//
//	"Choose one —
//	 • You gain 5 life.
//	 • Counter target spell.
//	 • Target creature gets -2/-2 until end of turn."
//
// A modal spell (CR 700.2) whose bullets are the catalog's mode bodies:
// Abuna's Chant's life for the caster, CounterTheModesTarget, and
// Auntie's Sentence's -2/-2. Each targeted bullet's target is checked
// against its own clause at announce and again at resolution
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e9950393-8458-4132-9dc7-01246282a41a",
		Name:         "Dromar's Charm",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeDoing("You gain 5 life.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.GainLife{Amount: 5}.Apply(ctx)
				}),
			effects.ModeDoing("Counter target spell.",
				effects.TargetSpell("target spell"),
				effects.CounterTheModesTarget),
			effects.ModeWithPurpose(effects.ModeDoing("Target creature gets -2/-2 until end of turn.",
				effects.TargetCreature("target creature"),
				effects.BoostTheModesTarget(-2, -2, "Dromar's Charm")), effects.ForTargets(effects.HarmUntilEOT(0, -2, -2))),
		),
	})
}
