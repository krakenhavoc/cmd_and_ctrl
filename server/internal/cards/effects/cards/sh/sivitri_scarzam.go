package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sivitri Scarzam — Legendary Creature — Human {5}{U}{B}, 6/4 (Legends, #5639)
//
// No rules text: a vanilla 6/4. Registered so the catalogue lists it
// as reviewed; there is nothing for the engine to do beyond its printed
// power and toughness.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "333cc91b-01ab-4491-a993-d8304fc382c6",
		Name:         "Sivitri Scarzam",
		Completeness: effects.CompletenessFull,
	})
}
