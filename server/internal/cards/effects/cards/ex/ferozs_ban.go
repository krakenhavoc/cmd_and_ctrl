package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Feroz's Ban — Artifact {6} (Homelands, S82 OLD b827, #5635):
//
//	"Creature spells cost {2} more to cast."
//
// Thorn of Amethyst's symmetric tax over creature spells and by {2}: it taxes
// its controller's creatures exactly as it taxes everyone else's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "445dd350-5165-4516-8669-9233f5558605",
		Name:         "Feroz's Ban",
		Completeness: effects.CompletenessFull,
		CostModifiers: []game.CostModifier{
			effects.CostsMore(2, "Creature spells cost {2} more to cast.", effects.CreatureSpell()),
		},
	})
}
