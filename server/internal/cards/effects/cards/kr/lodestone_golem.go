package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lodestone Golem — Artifact Creature — Golem {4}, 5/3:
//
//	"Nonartifact spells cost {1} more to cast."
//
// A symmetrical tax: it applies to every player's spells, its controller's
// included (CR 601.2f).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a698c9f5-325a-49f6-8902-eef09f4260c5",
		Name:         "Lodestone Golem",
		Completeness: effects.CompletenessFull,
		CostModifiers: []game.CostModifier{
			effects.CostsMore(1, "Nonartifact spells cost {1} more to cast.", effects.NonartifactSpell()),
		},
	})
}
