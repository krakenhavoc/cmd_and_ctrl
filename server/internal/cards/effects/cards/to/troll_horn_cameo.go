package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Troll-Horn Cameo — Artifact {3} (Invasion, #5616): "{T}: Add {R} or {G}."
//
// A mana ability (CR 605.1a) that offers its two colours, the Mox shape.
//
// No simplification.
func init() {
	effects.Register(effects.S82OldB657Cameo("98e042de-05f4-4e2e-b12f-375b905e6600", "Troll-Horn Cameo", "R", "G"))
}
