package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scabland — Land (Tempest, S82 OLD, #5628):
//
//	"This land enters tapped.
//	 {T}: Add {C}.
//	 {T}: Add {R} or {W}. This land deals 1 damage to you."
//
// The enters-tapped painland of the Tempest cycle: the painlands' two separate
// mana abilities (a free colourless half and a coloured half that always hurts,
// painlands.go) with the self enters-tapped replacement of the Bad River family.
// The damage is a rider, not a cost, so the land stays activatable at 1 life.
//
// No simplification.
func init() {
	effects.Register(effects.S82OldB763TappedPainland("2896f01d-003b-4d68-9d0a-64990ba59cbe", "Scabland", "R", "W"))
}
