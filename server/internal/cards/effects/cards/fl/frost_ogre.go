package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Frost Ogre — Creature — Ogre Warrior {3}{R}{R}, 5/3 (Betrayers of Kamigawa, S82 OLD
// b595): no rules text.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b039997c-2115-4f6d-9f7b-0ad35a4796a6",
		Name:         "Frost Ogre",
		Completeness: effects.CompletenessFull,
	})
}
