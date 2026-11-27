package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Seashell Cameo — Artifact {3} (Invasion, #5616): "{T}: Add {W} or {U}."
//
// A mana ability (CR 605.1a) that offers its two colours, the Mox shape.
//
// No simplification.
func init() {
	effects.Register(effects.S82OldB657Cameo("383f6020-c26d-43e8-bb07-566886626d74", "Seashell Cameo", "W", "U"))
}
