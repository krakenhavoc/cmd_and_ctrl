package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Deface — Sorcery {R} (Ravnica Allegiance, #5521):
//
//	"Choose one —
//	 • Destroy target artifact.
//	 • Destroy target creature with defender."
//
// Choose one (CR 700.2): Bant Charm's artifact bullet, and a creature
// bullet whose target must have defender (a layer-6 keyword, so a granted
// defender counts).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "390290b3-77cc-4180-8e6c-9cb808f6ac7a",
		Name:         "Deface",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeDoing("Destroy target artifact.",
				effects.TargetPermanent("target artifact", effects.Artifact()),
				effects.DestroyTheModesTarget),
			effects.ModeDoing("Destroy target creature with defender.",
				effects.TargetCreature("target creature with defender", effects.HasKeyword("defender")),
				effects.DestroyTheModesTarget),
		),
	})
}
