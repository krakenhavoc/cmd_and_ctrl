package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bloodstone Cameo — Artifact {3} (Invasion, #5616): "{T}: Add {B} or {R}."
//
// A mana ability (CR 605.1a) that offers its two colours, the Mox shape.
//
// No simplification.
func init() {
	effects.Register(effects.S82OldB657Cameo("1ce6ae30-33c3-4f05-9286-69b0871b1c2d", "Bloodstone Cameo", "B", "R"))
}
