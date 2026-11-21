package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Will-Forged Golem — Artifact Creature — Golem {6}, 4/4 (Magic 2015,
// #5556):
//
//	"Convoke (Your creatures can help cast this spell. Each creature you
//	 tap while casting this spell pays for {1} or one mana of that
//	 creature's color.)"
//
// Halo Hopper's convoke: the TapCost every convoke card declares.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "25fb0396-7355-4ff1-8ad3-d5a82adf6cc1",
		Name:         "Will-Forged Golem",
		Completeness: effects.CompletenessFull,
		TapCost:      effects.Convoke(),
	})
}
