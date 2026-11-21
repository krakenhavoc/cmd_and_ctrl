package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scoria Elemental — Creature — Elemental {4}{R}, 6/1 (Scars of Mirrodin, #5577)
//
// No rules text: a vanilla body. Registered so the catalogue lists it
// as reviewed; its cost, type line and power and toughness are printed
// card data.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1a6e8369-241f-4ddc-b4cd-a974d6b689a7",
		Name:         "Scoria Elemental",
		Completeness: effects.CompletenessFull,
	})
}
