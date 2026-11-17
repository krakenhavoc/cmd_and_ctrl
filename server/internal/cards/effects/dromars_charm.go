package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "e9950393-8458-4132-9dc7-01246282a41a",
		Name:         "Dromar's Charm",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeDoing("You gain 5 life.", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					return GainLife{Amount: 5}.Apply(ctx)
				}),
			ModeDoing("Counter target spell.",
				TargetSpell("target spell"),
				CounterTheModesTarget),
			ModeWithPurpose(ModeDoing("Target creature gets -2/-2 until end of turn.",
				TargetCreature("target creature"),
				BoostTheModesTarget(-2, -2, "Dromar's Charm")), ForTargets(HarmUntilEOT(0, -2, -2))),
		),
	})
}
