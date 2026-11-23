package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mardu Outrider — Creature — Orc Warrior {1}{B}{B}, 5/5 (Mystery Booster 2):
//
//	"As an additional cost to cast this spell, discard a card."
//
// The discard is an additional cost (CR 118.8) paid as the spell is cast,
// so a hand with no other card cannot cast it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "9bc5ed59-ae13-40df-94b9-c56c32673857",
		Name:           "Mardu Outrider",
		Completeness:   effects.CompletenessFull,
		AdditionalCost: effects.DiscardCost(1),
	})
}
