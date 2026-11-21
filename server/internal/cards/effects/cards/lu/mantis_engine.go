package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mantis Engine — Artifact Creature — Insect {5}, 3/3 (Urza's Destiny,
// #5621):
//
//	"{2}: This creature gains flying until end of turn.
//	 {2}: This creature gains first strike until end of turn."
//
// Two rows of the shared udsB3ThisGainsRow: each is a layer-6 grant on
// the Engine that ends in the cleanup step (CR 514.2). Neither cost has
// {T}, so both can be activated the turn it arrives.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "57bea272-5b38-444e-acc2-6104af2ece22",
		Name:         "Mantis Engine",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			effects.UdsB3ThisGainsRow("{2}", "flying", "Mantis Engine"),
			effects.UdsB3ThisGainsRow("{2}", "first strike", "Mantis Engine"),
		},
	})
}
