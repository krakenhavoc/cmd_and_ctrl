package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Root Cage — Enchantment {1}{G} (Prophecy, S82 OLD batch 674, #5617):
//
//	"Mercenaries don't untap during their controllers' untap steps."
//
// Back to Basics' untap-step restriction filtered to Mercenaries, whoever
// controls them. The creature type is read through the layered view, so a
// changeling is restricted.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:              "14c3a43c-cdb1-45d1-8eba-8e6d16bd7643",
		Name:                  "Root Cage",
		Completeness:          effects.CompletenessFull,
		UntapStepRestrictions: []game.UntapStepRestriction{effects.DoesntUntapDuringTheirControllersUntapSteps(effects.Subtype("Mercenary"))},
	})
}
