package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thorn of Amethyst — Artifact {2}:
//
//	"Noncreature spells cost {1} more to cast."
//
// Sphere of Resistance with a type predicate. Still symmetrical — it
// taxes its controller's removal spell exactly as it taxes everyone
// else's.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0c6c5336-9233-4ab1-9d55-79f20be7ea57",
		Name:         "Thorn of Amethyst",
		Completeness: effects.CompletenessFull,
		CostModifiers: []game.CostModifier{
			effects.CostsMore(1, "Noncreature spells cost {1} more to cast.", effects.NoncreatureSpell()),
		},
	})
}
