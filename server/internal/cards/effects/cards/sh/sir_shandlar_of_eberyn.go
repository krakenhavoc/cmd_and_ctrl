package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sir Shandlar of Eberyn — Legendary Creature — Human Knight {4}{G}{W}, 4/7 (Legends, #5639)
//
// No rules text: a vanilla 4/7. Registered so the catalogue lists it
// as reviewed; there is nothing for the engine to do beyond its printed
// power and toughness.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7a9942cd-32ac-4cc9-8840-3fe6b092d0b8",
		Name:         "Sir Shandlar of Eberyn",
		Completeness: effects.CompletenessFull,
	})
}
