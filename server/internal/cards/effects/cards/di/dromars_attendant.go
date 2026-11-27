package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dromar's Attendant — Artifact Creature — Golem {5}, 3/3 (Invasion, #5616):
//
//	"{1}, Sacrifice this creature: Add {W}{U}{B}."
//
// A mana ability (CR 605.1a): the cost pays {1} and sacrifices the Golem,
// and the three colours come out together. No {T}, so summoning sickness
// does not stop it.
//
// No simplification.
func init() {
	effects.Register(effects.S82OldB657Attendant("91d36a9d-3a3f-4cb7-837d-b1e8f6730512", "Dromar's Attendant", "{W}{U}{B}"))
}
