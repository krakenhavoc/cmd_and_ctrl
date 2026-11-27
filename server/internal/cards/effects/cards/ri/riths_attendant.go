package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rith's Attendant — Artifact Creature — Golem {5}, 3/3 (Invasion, #5616):
//
//	"{1}, Sacrifice this creature: Add {R}{G}{W}."
//
// A mana ability (CR 605.1a): the cost pays {1} and sacrifices the Golem,
// and the three colours come out together. No {T}, so summoning sickness
// does not stop it.
//
// No simplification.
func init() {
	effects.Register(effects.S82OldB657Attendant("781542b7-155b-4a66-aa52-eec0ebb29bb2", "Rith's Attendant", "{R}{G}{W}"))
}
