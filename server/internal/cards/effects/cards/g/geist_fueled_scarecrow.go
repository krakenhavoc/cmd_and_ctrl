package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Geist-Fueled Scarecrow — {4} Artifact Creature — Scarecrow 4/4 (Eldritch
// Moon, S82 OLD, #5541):
//
//	"Creature spells you cast cost {1} more to cast."
//
// A tax on its own controller's creature spells only (CostsMore with
// YourSpell), while it is on the battlefield.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b925c4ed-8767-46b8-a423-a5f094557d94",
		Name:         "Geist-Fueled Scarecrow",
		Completeness: effects.CompletenessFull,
		CostModifiers: []game.CostModifier{
			effects.CostsMore(1, "Creature spells you cast cost {1} more to cast.", effects.YourSpell(), effects.CreatureSpell()),
		},
	})
}
