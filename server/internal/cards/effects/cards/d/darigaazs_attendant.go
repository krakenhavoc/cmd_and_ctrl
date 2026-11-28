package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Darigaaz's Attendant — Artifact Creature — Golem {5}, 3/3 (Invasion, #5616):
//
//	"{1}, Sacrifice this creature: Add {B}{R}{G}."
//
// A mana ability (CR 605.1a): the cost pays {1} and sacrifices the Golem,
// and the three colours come out together. No {T}, so summoning sickness
// does not stop it.
//
// No simplification.
func init() {
	effects.Register(effects.S82OldB657Attendant("c0ad16d7-2550-47e0-8581-0e22b27cb0d0", "Darigaaz's Attendant", "{B}{R}{G}"))
}
