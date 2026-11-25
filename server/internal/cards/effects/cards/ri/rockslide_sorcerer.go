package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rockslide Sorcerer — Creature — Human Wizard Sorcerer {3}{R}, 3/3
// (Zendikar Rising, S82 OLD #5508):
//
//	"Whenever you cast an instant, sorcery, or Wizard spell, this creature
//	 deals 1 damage to any target."
//
// The target is chosen as the trigger goes on the stack (CR 601.2i,
// 603.3d).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d8b7b78e-91b2-4ad5-aeb8-6ebdca363e71",
		Name:         "Rockslide Sorcerer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WheneverYouCast(effects.Or(effects.Instant(), effects.Sorcery(), effects.HasSubtype("Wizard")),
					"Rockslide Sorcerer — it deals 1 damage to any target",
					effects.B33DamageChosenTargetFromSource(1)),
				effects.TargetAny()),
		},
	})
}
