package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Treva's Attendant — Artifact Creature — Golem {5}, 3/3 (Invasion, #5616):
//
//	"{1}, Sacrifice this creature: Add {G}{W}{U}."
//
// A mana ability (CR 605.1a): the cost pays {1} and sacrifices the Golem,
// and the three colours come out together. No {T}, so summoning sickness
// does not stop it.
//
// No simplification.
func init() {
	effects.Register(effects.S82OldB657Attendant("ff7d927f-0a01-4243-ad00-c65686ee86bb", "Treva's Attendant", "{G}{W}{U}"))
}
