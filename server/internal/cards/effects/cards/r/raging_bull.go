package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Raging Bull — Creature — Ox {2}{R}, 2/2 (Legends, #5639)
//
// No rules text: a vanilla 2/2. Registered so the catalogue lists it
// as reviewed; there is nothing for the engine to do beyond its printed
// power and toughness.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "341680e2-12a8-484f-be09-24dae9dbfaea",
		Name:         "Raging Bull",
		Completeness: effects.CompletenessFull,
	})
}
