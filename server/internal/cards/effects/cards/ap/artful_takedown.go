package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Artful Takedown — Instant {2}{U}{B} (Guilds of Ravnica, #5525):
//
//	"Choose one or both —
//	 • Tap target creature.
//	 • Target creature gets -2/-4 until end of turn."
//
// Saheeli's Artistry's "one or both" modes, each with its own target (they may
// be the same creature): the chosen modes run in printed order, each against
// its own pick, and a pick that has become illegal gets nothing while the other
// mode still does (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "edc193bf-2987-4c7b-9cf3-69089ff13764",
		Name:         "Artful Takedown",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseN("Choose one or both", 1, 2,
			effects.ModeDoing("Tap target creature.",
				effects.TargetCreature("target creature"),
				func(_ *game.StackItem, ctx *effects.Context, occ int) error {
					t, ok := effects.ModeTarget(ctx, occ)
					if !ok || t.Kind != game.TargetCard {
						return nil
					}
					return effects.TapTarget{Target: t.ID}.Apply(ctx)
				}),
			effects.ModeWithPurpose(effects.ModeDoing("Target creature gets -2/-4 until end of turn.",
				effects.TargetCreature("target creature"),
				effects.BoostTheModesTarget(-2, -4, "Artful Takedown — -2/-4")), effects.ForTargets(effects.HarmUntilEOT(0, -2, -4))),
		),
	})
}
