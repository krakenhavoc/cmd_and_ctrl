package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Destructive Tampering — Sorcery {2}{R} (Jumpstart 2022, #5948):
//
//	"Choose one —
//	 • Destroy target artifact.
//	 • Creatures without flying can't block this turn."
//
// Abrade's destroy-an-artifact bullet, and Falter's restriction (faltering)
// as the second bullet, which targets nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e6ae536a-95c0-495d-98b8-200b3564dd79",
		Name:         "Destructive Tampering",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeDoing("Destroy target artifact.",
				effects.TargetPermanent("target artifact", effects.Artifact()),
				effects.DestroyTheModesTarget),
			effects.ModeDoing("Creatures without flying can't block this turn.", nil,
				func(item *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.Faltering("Destructive Tampering")(item, ctx)
				}),
		),
	})
}
