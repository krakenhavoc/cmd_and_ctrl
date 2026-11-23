package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lady Orca — Legendary Creature — Demon {5}{B}{R}, 7/4 (Legends, #5639)
//
// No rules text: a vanilla 7/4. Registered so the catalogue lists it
// as reviewed; there is nothing for the engine to do beyond its printed
// power and toughness.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fcae9e87-50af-46b4-ab48-42ed92985e10",
		Name:         "Lady Orca",
		Completeness: effects.CompletenessFull,
	})
}
