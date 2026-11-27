package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hum of the Radix — Enchantment {2}{G}{G} (Mirrodin, #5606):
//
//	"Each artifact spell costs {1} more to cast for each artifact its
//	 controller controls."
//
// Damping Sphere's scaling cost increase (CR 601.2f), over every artifact spell
// whoever casts it, with the amount the artifacts the caster controls at the
// time the cost is worked out.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "95803a5a-9f98-4764-b6bd-6cf4cdcb740b",
		Name:         "Hum of the Radix",
		Completeness: effects.CompletenessFull,
		CostModifiers: []game.CostModifier{
			effects.CostsMoreEach(effects.OldB607ArtifactsYouControl,
				"Each artifact spell costs {1} more to cast for each artifact its controller controls.", effects.ArtifactSpell()),
		},
	})
}
