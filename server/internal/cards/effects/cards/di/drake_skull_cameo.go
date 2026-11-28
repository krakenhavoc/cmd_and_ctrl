package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Drake-Skull Cameo — Artifact {3} (Invasion, #5616): "{T}: Add {U} or {B}."
//
// A mana ability (CR 605.1a) that offers its two colours, the Mox shape.
//
// No simplification.
func init() {
	effects.Register(effects.S82OldB657Cameo("8fbdec25-4222-4b96-aeea-82a4b5b8b80e", "Drake-Skull Cameo", "U", "B"))
}
