package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jasmine Boreal — Legendary Creature — Human {3}{G}{W}, 4/5 (Legends, #5639)
//
// No rules text: a vanilla 4/5. Registered so the catalogue lists it
// as reviewed; there is nothing for the engine to do beyond its printed
// power and toughness.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "942407bf-862a-4a0b-b19d-b9f8ee833026",
		Name:         "Jasmine Boreal",
		Completeness: effects.CompletenessFull,
	})
}
