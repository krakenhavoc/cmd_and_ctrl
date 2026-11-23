package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tobias Andrion — Legendary Creature — Human Advisor {3}{W}{U}, 4/4 (Legends, #5639)
//
// No rules text: a vanilla 4/4. Registered so the catalogue lists it
// as reviewed; there is nothing for the engine to do beyond its printed
// power and toughness.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "55dfe722-7487-4d25-b90c-e6ad2c82a671",
		Name:         "Tobias Andrion",
		Completeness: effects.CompletenessFull,
	})
}
