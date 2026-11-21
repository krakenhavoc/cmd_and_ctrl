package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vivid Marsh — Land (Vivid cycle, Lorwyn, #5593):
//
//	"This land enters tapped with two charge counters on it.
//	 {T}: Add {B}.
//	 {T}, Remove a charge counter from this land: Add one mana of any
//	 color."
//
// The cycle's shared vividLand (vivid_lands.go).
//
// No simplification.
func init() {
	effects.Register(effects.VividLand("20b32052-f66f-4eb8-b56e-00d531907f19", "Vivid Marsh", "B"))
}
