package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lurching Rotbeast — Creature — Zombie Beast {3}{B}, 4/2
// (Hour of Devastation, #5534):
//
//	"Cycling {B} ({B}, Discard this card: Draw a card.)"
//
// The keywords are printed data; cycling is the shared Cycling row, a
// hand ability that discards the card as its cost and draws (CR 702.29a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "771be8ad-5ab5-40b1-ad9b-e5e834949d32",
		Name:         "Lurching Rotbeast",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.Cycling("{B}")},
	})
}
